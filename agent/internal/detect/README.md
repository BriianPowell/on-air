# Detection

OS-level mic/camera detection for the on-air agent.

```
detect/
├── types/           # Status + Detector interface (shared, no platform imports)
├── platform/
│   ├── darwin/      # macOS — CoreAudio + CoreMediaIO (CGO)
│   └── windows/     # Windows — WASAPI + registry camera heuristic
├── detect.go        # Re-exports types for callers
└── new_*.go         # Platform factory (build tags) — use detect.New()
```

Callers import `github.com/brianpowell/on-air/internal/detect` and call `detect.New()`.
