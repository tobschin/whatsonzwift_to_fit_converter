# WhatsonZwift to FIT Converter

Command-line tool that converts workouts from [whatsonzwift.com](https://whatsonzwift.com) into `.fit` files. The generated FIT files use **relative power (% FTP)** – the actual wattage is calculated by your device (Garmin, Wahoo, etc.) based on your configured FTP.

## Prerequisites

- [Go](https://go.dev/dl/) ≥ 1.21

## Build

### Mac / Linux

```bash
go build -o whatsonzwift-to-fit .
```

### Windows (PowerShell)

```powershell
go build -o whatsonzwift-to-fit.exe .
```

### Cross-Compilation (on Mac for other platforms)

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o whatsonzwift-to-fit-linux .

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o whatsonzwift-to-fit.exe .
```

## Usage

```bash
./whatsonzwift-to-fit <whatsonzwift-workout-url>
```

### Example: "The Gorby"

```bash
./whatsonzwift-to-fit https://whatsonzwift.com/workouts/less-than-60-minutes-to-burn/the-gorby
```

**Output:**

```
Fetching workout from: https://whatsonzwift.com/workouts/less-than-60-minutes-to-burn/the-gorby
Workout: The Gorby
Steps parsed: 13
  Warmup: 60s 30%→35% FTP
  Warmup: 60s 35%→40% FTP
  Warmup: 60s 40%→45% FTP
  Warmup: 60s 45%→50% FTP
  Warmup: 60s 50%→55% FTP
  Warmup: 60s 55%→60% FTP
  Warmup: 60s 60%→65% FTP
  Warmup: 60s 65%→70% FTP
  Warmup: 60s 70%→75% FTP
  Warmup: 60s 75%→80% FTP
  Active: 300s 110%→110% FTP
  Rest: 300s @ 55% FTP
  Repeat 5x
FIT steps: 13
Written: output/fit/the_gorby.fit (396 bytes)
Written: output/md/the_gorby.md
```

Two files are generated in subfolders:
- **`output/fit/the_gorby.fit`** – FIT workout file for import on Garmin/Wahoo/TrainingPeaks
- **`output/md/the_gorby.md`** – Markdown summary with ASCII art and interval table

### Example: Generated Markdown (`the_gorby.md`)

````markdown
# The Gorby

```
110% |          #####     #####     #####     #####     #####     
101% |          #####     #####     #####     #####     #####     
 92% |          #####     #####     #####     #####     #####     
 83% |          #####     #####     #####     #####     #####     
 74% |         ######     #####     #####     #####     #####     
 65% |       ########     #####     #####     #####     #####     
 55% |     #######################################################
 46% |   #########################################################
 37% | ###########################################################
 28% |############################################################
 19% |############################################################
 10% |############################################################
     +------------------------------------------------------------
      0 min                                                  60 min
```

## Intervals

| # | Phase | Duration | % FTP |
|---|-------|----------|-------|
| 1 | Warmup | 10.0 min | 30% → 80% |
| 2 | Active | 5.0 min | 110% |
| 3 | Rest | 5.0 min | 55% |
| 4 | **Repeat** | — | **5×** |

**Total duration:** 60 min
````

### Windows

```powershell
.\whatsonzwift-to-fit.exe https://whatsonzwift.com/workouts/less-than-60-minutes-to-burn/the-gorby
```

## Workout Structure (Example: The Gorby)

| Phase | Duration | Intensity |
|-------|----------|----------|
| Warmup (Ramp) | 10 min | 30% → 80% FTP |
| Interval (5×) | 5 min | 110% FTP |
| Rest (5×) | 5 min | 55% FTP |

**Total:** 60 minutes, 81 Stress Points

## How does it work?

1. The URL is fetched and the workout page HTML is parsed
2. The workout description (e.g. `10min from 30 to 80% FTP`, `5x 5min @ 110% FTP, 5min @ 55% FTP`) is split into interval steps
3. Ramps are divided into 1-minute steps
4. The steps are encoded as a FIT workout file with relative power targets (% FTP)

## Supported Workout Formats

- **Ramps:** `10min from 30 to 80% FTP`
- **Steady-State:** `5min @ 75% FTP`
- **Repeats:** `5x 5min @ 110% FTP, 5min @ 55% FTP`
- **Seconds Intervals:** `30s @ 150% FTP`

## Project Structure

```
.
├── main.go              # CLI entry point
├── scraper/
│   └── scraper.go       # HTML scraper for whatsonzwift.com
├── fit/
│   └── encoder.go       # FIT file encoder (binary format)
├── output/              # Generated files (ignored via .gitignore)
│   ├── fit/             # .fit workout files
│   └── md/              # .md summaries
├── go.mod
├── go.sum
└── README.md
```

## License

MIT
