---
applyTo: "**"
---

## Technology Stack
| Component | Technology         | Version |
|-----------|--------------------|---------|
| Backend   | GoLang             | 1.26   |

## Language Settings
README.md and output in English

## Supported Time Formats
| Format | Example | Description |
|--------|---------|-------------|
| `Xmin` | `5min`, `15min` | Duration in minutes |
| `Xs`   | `30s` | Duration in seconds |
| `Xsec` | `30sec` | Duration in seconds (alternative notation) |

## Supported Workout Description Formats
| Pattern | Example | Description |
|---------|---------|-------------|
| Steady-State | `5min @ 110% FTP` | Constant power |
| Seconds Interval | `30sec @ 109% FTP` | Short interval |
| Ramp | `10min from 30 to 80% FTP` | Linear ramp up/down |
| Repeat | `5x 5min @ 110% FTP, 5min @ 55% FTP` | Repeated interval blocks |
| Cadence + Ramp | `15min @ 65rpm, from 65 to 92% FTP` | Cadence is ignored, ramp is parsed |
| Free Ride | `2min free ride` | Free riding → 70% FTP |
| Decimal FTP values | `109.44999` | Rounded to whole numbers |

## Program Behavior
- Free ride segments are treated with **70% FTP** as target power
- Cadence annotations (`@ Xrpm,`) are stripped before parsing
- Ramps are split into **1-minute steps**
- Each step gets a **±5% FTP buffer** as target range in the FIT file
- Power targets are relative (% FTP), no absolute watt value needed