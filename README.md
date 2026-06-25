# WhatsonZwift to FIT Converter

Kommandozeilen-Tool, das Workouts von [whatsonzwift.com](https://whatsonzwift.com) in `.fit`-Dateien konvertiert. Die erzeugten FIT-Dateien arbeiten mit **relativer Leistung (% FTP)** – die tatsächliche Wattzahl wird vom Gerät (Garmin, Wahoo etc.) anhand deiner konfigurierten FTP berechnet.

## Voraussetzungen

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

### Cross-Compilation (auf Mac für andere Plattformen)

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o whatsonzwift-to-fit-linux .

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o whatsonzwift-to-fit.exe .
```

## Verwendung

```bash
./whatsonzwift-to-fit <whatsonzwift-workout-url>
```

### Beispiel: "The Gorby"

```bash
./whatsonzwift-to-fit https://whatsonzwift.com/workouts/less-than-60-minutes-to-burn/the-gorby
```

**Ausgabe:**

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

Es werden zwei Dateien in Unterordnern erzeugt:
- **`output/fit/the_gorby.fit`** – FIT-Workout-Datei zum Import auf Garmin/Wahoo/TrainingPeaks
- **`output/md/the_gorby.md`** – Markdown-Zusammenfassung mit ASCII-Art und Intervall-Tabelle

### Beispiel: Generiertes Markdown (`the_gorby.md`)

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

## Intervalle

| # | Phase | Dauer | % FTP |
|---|-------|-------|-------|
| 1 | Warmup | 10.0 min | 30% → 80% |
| 2 | Active | 5.0 min | 110% |
| 3 | Rest | 5.0 min | 55% |
| 4 | **Repeat** | — | **5×** |

**Gesamtdauer:** 60 min
````

### Windows

```powershell
.\whatsonzwift-to-fit.exe https://whatsonzwift.com/workouts/less-than-60-minutes-to-burn/the-gorby
```

## Workout-Struktur (Beispiel: The Gorby)

| Phase | Dauer | Intensität |
|-------|-------|------------|
| Warmup (Rampe) | 10 min | 30% → 80% FTP |
| Intervall (5×) | 5 min | 110% FTP |
| Pause (5×) | 5 min | 55% FTP |

**Gesamt:** 60 Minuten, 81 Stress Points

## Wie funktioniert es?

1. Die URL wird aufgerufen und das HTML der Workout-Seite geparst
2. Die Workout-Beschreibung (z.B. `10min from 30 to 80% FTP`, `5x 5min @ 110% FTP, 5min @ 55% FTP`) wird in Intervall-Schritte zerlegt
3. Rampen werden in 1-Minuten-Schritte aufgeteilt
4. Die Schritte werden als FIT Workout-Datei mit relativen Leistungszielen (% FTP) kodiert

## Unterstützte Workout-Formate

- **Rampen:** `10min from 30 to 80% FTP`
- **Steady-State:** `5min @ 75% FTP`
- **Wiederholungen:** `5x 5min @ 110% FTP, 5min @ 55% FTP`
- **Sekunden-Intervalle:** `30s @ 150% FTP`

## Projektstruktur

```
.
├── main.go              # CLI-Einstiegspunkt
├── scraper/
│   └── scraper.go       # HTML-Scraper für whatsonzwift.com
├── fit/
│   └── encoder.go       # FIT-Datei-Encoder (Binärformat)
├── output/              # Generierte Dateien (via .gitignore ignoriert)
│   ├── fit/             # .fit Workout-Dateien
│   └── md/              # .md Zusammenfassungen
├── go.mod
├── go.sum
└── README.md
```

## Lizenz

MIT
