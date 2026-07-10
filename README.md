# Voice Pipeline — English → Arabic (Go)

Real ASR → Translation → TTS pipeline. Not pseudocode — this hits real APIs
and returns playable Arabic audio.

## 1. What you need to install

| Requirement | Why |
|---|---|
| **Go 1.21+** | to build/run the server |
| **Groq API key** | covers Translation (Mixtral) — Groq supports chat/text only |
| **curl** or **Postman** | to send a test audio file to the server |
| Any audio player | to play the returned `arabic_output.mp3` |

You do **not** need ffmpeg. Groq's API accepts `wav`, `mp3`, `m4a` directly.

```bash
export GROQ_API_KEY="gsk-...your-key..."
```

## 2. Get a sample English voice input

You need one short English audio file to test with. Easiest options:

- **Record it yourself**: any phone voice-memo app, say "Can we move the
  meeting to 5 PM tomorrow?", export as `sample.wav` or `sample.m4a`.
- **Or generate one**: use any free online text-to-speech tool, type the
  sentence above, download the audio, name it `sample.wav`.

Put it in the project root: `VA/sample.wav`.

## 3. Run it

```bash
cd VA
go mod tidy       # fetches github.com/google/uuid
go run cmd/main.go
```

You should see structured JSON logs like:

```json
{"time":"...","level":"INFO","msg":"startup.server_started","addr":":8080"}
```

## 4. Send your sample audio

```bash
curl -X POST http://localhost:8080/translate \
  -F "audio=@sample.wav" \
  -o arabic_output.mp3
```

Play `arabic_output.mp3` — that's the Arabic voice output.

While it runs, your terminal will show production-style logs for every stage:

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

Every stage is logged separately with the session ID, so you can point to
exactly where a failure happened — this is what "production-style logging"
means in practice, not just `fmt.Println`.

## 5. Run the tests

```bash
go test ./... -v
```

These use hand-written mocks (`mockTranscriber`, `mockTranslator`,
`mockSynthesizer`) — no real API calls in tests, no API key needed to run
them. This is the payoff of coding to interfaces (DIP): the pipeline logic
is fully testable without touching a network.

## 6. Is this "real-time"?

Honest answer: this is **near-real-time, per-utterance**, not continuous
streaming. You send one complete audio file, get one complete Hindi audio
file back — round trip is typically 2–4 seconds depending on sentence
length and network latency to OpenAI.

**True real-time streaming** (audio streamed in, partial Arabic audio
streamed out while you're still talking) needs:
- A WebSocket instead of a single POST endpoint
- Streaming ASR (partial transcripts as you speak)
- A VAD (voice-activity-detector) to know when you've finished a sentence
- Streaming TTS that starts playing before the full sentence is synthesized

That's the `pipeline`/`VADService` layer from the fuller architecture we
discussed earlier — worth mentioning in your assignment writeup as "future
work," but not necessary to get working Hindi audio out today.

## 7. Project structure recap

```
cmd/main.go                          → DI root, wires everything together
internal/model/                      → plain structs, no dependencies
internal/iface/                      → Transcriber, Translator, Synthesizer, SessionRepository
internal/provider/                   → real implementations (Groq chat/translation) + Factory
internal/repository/                 → in-memory session store
internal/service/                    → PipelineService (orchestration) + TextCleaner
internal/routes/                     → HTTP handler
internal/service/*_test.go           → unit tests with mocks
```
