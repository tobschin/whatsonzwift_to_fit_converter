---
applyTo: "**"
---

## Technologie Stack
| Komponente | Technologie        | Version |
|------------|--------------------|---------|
| Backend    | GoLang             | 1.26   |

## Unterstützte Zeitformate
| Format | Beispiel | Beschreibung |
|--------|----------|--------------|
| `Xmin` | `5min`, `15min` | Dauer in Minuten |
| `Xs`   | `30s` | Dauer in Sekunden |
| `Xsec` | `30sec` | Dauer in Sekunden (alternative Schreibweise) |

## Unterstützte Workout-Beschreibungsformate
| Muster | Beispiel | Beschreibung |
|--------|----------|--------------|
| Steady-State | `5min @ 110% FTP` | Konstante Leistung |
| Sekunden-Intervall | `30sec @ 109% FTP` | Kurzes Intervall |
| Rampe | `10min from 30 to 80% FTP` | Lineare Steigerung/Senkung |
| Wiederholung | `5x 5min @ 110% FTP, 5min @ 55% FTP` | Wiederholte Intervall-Blöcke |
| Kadenz + Rampe | `15min @ 65rpm, from 65 to 92% FTP` | Kadenz-Angabe wird ignoriert, Rampe geparst |
| Free Ride | `2min free ride` | Freies Fahren → 70% FTP |
| Dezimale FTP-Werte | `109.44999` | Werden auf ganze Zahlen gerundet |

## Programmverhalten
- Free-Ride-Sektoren werden mit **70% FTP** als Zielleistung behandelt
- Kadenz-Angaben (`@ Xrpm,`) werden vor dem Parsen entfernt
- Rampen werden in **1-Minuten-Schritte** zerlegt
- Jeder Step bekommt einen **±5% FTP Puffer** als Zielbereich in der FIT-Datei
- Leistungsziele sind relativ (% FTP), kein absoluter Watt-Wert nötig