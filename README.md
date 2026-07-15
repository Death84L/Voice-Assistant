# Voice Pipeline — English → Arabic (Go)

Real ASR → Translation → TTS pipeline. This hits real APIs and returns
playable Arabic audio. Supports both a **batch** endpoint (upload one
file, wait, get one file back) and a **streaming** endpoint (WebSocket,
short audio bursts in, translated audio out roughly every few seconds
while the connection is open).

## 1. What we need to install

| Requirement | Why |
|---|---|
| **Go 1.21+** | to build/run the server |
| **Groq API key** | covers ASR (Whisper), Translation (Llama), and TTS |
| **curl** or **Postman** | to test the batch endpoint |
| A modern browser | to test the streaming endpoint via `test_client.html` (mic access + WebSocket) |
| Any audio player | to play the returned `arabic_output.mp3` |

We do **not** need ffmpeg. Groq's API accepts `wav`, `mp3`, `m4a`, `webm`,
and `ogg` directly.

```bash
export GROQ_API_KEY="gsk-...your-key..."
```

## 2. Get a sample English voice input

We need one short English audio file to test the batch endpoint with.
Easiest options:

- **Record it**: any phone voice-memo app, say "Can we move the
  meeting to 5 PM tomorrow?", export as `sample.wav` or `sample.m4a`.
- **Or generate one**: use any free online text-to-speech tool, type the
  sentence above, download the audio, name it `sample.wav`.

Put it in the project root: `VA/sample.wav`.

For the streaming endpoint you don't need a sample file — `test_client.html`
records live from your microphone.

## 3. Run it

```bash
cd VA
go get github.com/gorilla/websocket   # only needed once, for streaming
go mod tidy
go run cmd/main.go
```

We should see structured JSON logs like:

```json
{"time":"...","level":"INFO","msg":"startup.server_started","addr":":8080"}
```

## 4a. Batch endpoint — send one file, get one file back

```bash
curl -X POST http://localhost:8080/translate \
  -F "audio=@sample.wav" \
  -o arabic_output.mp3
```

Play `arabic_output.mp3` — that's the Arabic voice output.

While it runs, the terminal shows structured logs for every stage:

```json
{"level":"INFO","msg":"http.request_received","session_id":"...","filename":"sample.wav","size_bytes":48210}
{"level":"INFO","msg":"asr.start","filename":"sample.wav"}
{"level":"INFO","msg":"asr.success","text":"Can we move the meeting to 5 PM tomorrow?"}
{"level":"INFO","msg":"pipeline.cleaned_text","text":"Can we move the meeting to 5 PM tomorrow?"}
{"level":"INFO","msg":"translate.start","input":"Can we move the meeting to 5 PM tomorrow?"}
{"level":"INFO","msg":"translate.success","output":"هل يمكننا تحديد اجتماعنا الساعة 5 مساء غدا؟"}
{"level":"INFO","msg":"tts.start","text":"هل يمكننا تحديد اجتماعنا الساعة 5 مساء غدا؟"}
{"level":"INFO","msg":"tts.success","audio_bytes":123456}
{"level":"INFO","msg":"pipeline.complete","session_id":"...","duration_ms":2140}
```

Every stage is logged separately with the session ID, so we can point to
exactly where a failure happened.

## 4b. Streaming endpoint — near-real-time over WebSocket

```
ws://localhost:8080/translate/stream
```

Open `test_client.html` in a browser, click **Start**, allow microphone
access, and talk. The client records in repeated 3-second bursts, sends
each burst over the WebSocket as it finishes, and plays back the
translated Arabic audio for each burst as it arrives — so you get
continuous back-and-forth without uploading one long file and waiting
for the whole thing to process.

Server-side, each burst is treated as an independent `AudioChunk` and run
through the exact same `PipelineService.Process()` used by the batch
endpoint — no duplicated ASR/translation/TTS logic between the two
endpoints. Logs look the same as above, once per chunk, tagged with the
same `session_id` for the life of the WebSocket connection.

## 5. Run the tests

```bash
go test ./... -v
```

These use hand-written mocks (`mockTranscriber`, `mockTranslator`,
`mockSynthesizer`) — no real API calls in tests, no API key needed to run
them. This is the payoff of coding to interfaces (DIP): the pipeline logic
is fully testable without touching a network.

## 6. Is this "real-time"?

**Batch endpoint (`/translate`)**: no — one file in, one file out, no
sense of "live."

**Streaming endpoint (`/translate/stream`)**: near-real-time,
chunk-by-chunk. You don't wait for a full recording to finish before
anything happens — translated audio starts coming back every ~3 seconds
while you're still talking. Round trip per chunk is typically 2–4 seconds
depending on chunk length and Groq API latency.

What this deliberately does **not** do, scoped out for the assignment:

- No VAD (voice-activity detection) — chunking is fixed-duration (3s),
  not silence/sentence-boundary aware
- No partial/interim ASR — each chunk is transcribed as a complete
  mini-utterance, not word-by-word
- No reconnect/resume — a dropped WebSocket just ends the session
- No horizontal scaling — session state is still an in-memory map, fine
  for one process

True sub-second simultaneous-interpretation-style streaming (partial
transcripts while you're mid-sentence, streaming TTS starting before
translation of the full sentence finishes) needs a streaming-capable ASR
provider (Groq's Whisper endpoint is batch-only) plus VAD plus
partial/final transcript handling — worth mentioning in the assignment
writeup as "future work," but out of scope here.

## 7. Project structure recap

```
cmd/main.go                          → DI root, wires everything together
internal/model/                      → plain structs, no dependencies
internal/iface/                      → Transcriber, Translator, Synthesizer, SessionRepository
internal/provider/                   → real implementations (Groq ASR/translation/TTS) + Factory
internal/repository/                 → in-memory session store
internal/service/                    → PipelineService (orchestration) + TextCleaner
internal/routes/                     → HTTP handler (batch) + WebSocket handler (streaming)
internal/service/*_test.go           → unit tests with mocks
test_client.html                     → browser test client for the streaming endpoint
```
