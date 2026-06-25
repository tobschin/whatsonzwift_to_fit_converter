package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/whatsonzwift-to-fit/fit"
	"github.com/whatsonzwift-to-fit/scraper"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <whatsonzwift-url>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s https://whatsonzwift.com/workouts/less-than-60-minutes-to-burn/the-gorby\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]

	if !strings.Contains(url, "whatsonzwift.com/workouts") {
		fmt.Fprintf(os.Stderr, "Error: URL must be a whatsonzwift.com workout link\n")
		os.Exit(1)
	}

	fmt.Printf("Fetching workout from: %s\n", url)

	workout, err := scraper.FetchWorkout(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Workout: %s\n", workout.Name)
	fmt.Printf("Steps parsed: %d\n", len(workout.Steps))

	for _, s := range workout.Steps {
		switch s.Type {
		case scraper.StepRepeat:
			fmt.Printf("  Repeat %dx\n", s.RepeatCount)
		case scraper.StepWarmup:
			fmt.Printf("  Warmup: %ds %d%%→%d%% FTP\n", s.DurationSec, s.PowerStartPct, s.PowerEndPct)
		case scraper.StepCooldown:
			fmt.Printf("  Cooldown: %ds %d%%→%d%% FTP\n", s.DurationSec, s.PowerStartPct, s.PowerEndPct)
		case scraper.StepRest:
			fmt.Printf("  Rest: %ds @ %d%% FTP\n", s.DurationSec, s.PowerStartPct)
		default:
			fmt.Printf("  Active: %ds %d%%→%d%% FTP\n", s.DurationSec, s.PowerStartPct, s.PowerEndPct)
		}
	}

	// Convert scraper steps to FIT steps
	fitSteps := convertToFitSteps(workout.Steps)

	fmt.Printf("FIT steps: %d\n", len(fitSteps))

	// Encode FIT file
	encoder := fit.NewEncoder()
	data, err := encoder.Encode(workout.Name, fitSteps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding FIT: %v\n", err)
		os.Exit(1)
	}

	// Generate output directory per workout
	baseName := generateBaseName(url)
	workoutDir := filepath.Join("output", baseName)
	if err := os.MkdirAll(workoutDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Write files into workout directory
	fitPath := filepath.Join(workoutDir, baseName+".fit")
	mdPath := filepath.Join(workoutDir, baseName+".md")

	err = os.WriteFile(fitPath, data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Written: %s (%d bytes)\n", fitPath, len(data))

	// Generate markdown summary
	mdContent := generateMarkdown(workout)
	err = os.WriteFile(mdPath, []byte(mdContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing markdown: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Written: %s\n", mdPath)
}

func convertToFitSteps(steps []scraper.WorkoutStep) []fit.Step {
	var fitSteps []fit.Step

	i := 0
	for i < len(steps) {
		step := steps[i]

		switch step.Type {
		case scraper.StepRepeat:
			// Find the sub-steps before this repeat marker
			// The repeat applies to the steps immediately before it
			// Count how many non-repeat steps precede this repeat
			subStepCount := 0
			for j := i - 1; j >= 0; j-- {
				if steps[j].Type == scraper.StepRepeat {
					break
				}
				subStepCount++
			}
			// The repeat step references the first step in the repeated block
			repeatFromIdx := uint32(len(fitSteps) - subStepCount)
			fitSteps = append(fitSteps, fit.Step{
				IsRepeat:      true,
				DurationValue: repeatFromIdx,
				TargetValue:   uint32(step.RepeatCount),
			})

		default:
			intensity := stepTypeToIntensity(step.Type)
			// Duration in milliseconds
			durationMs := uint32(step.DurationSec) * 1000

			// Power targets as % FTP
			// Average of start/end for the target
			lowPct := uint32(step.PowerStartPct)
			highPct := uint32(step.PowerEndPct)

			fitSteps = append(fitSteps, fit.Step{
				IsRepeat:         false,
				DurationValue:    durationMs,
				CustomTargetLow:  lowPct,
				CustomTargetHigh: highPct,
				Intensity:        intensity,
			})
		}
		i++
	}

	return fitSteps
}

func stepTypeToIntensity(t scraper.StepType) uint8 {
	switch t {
	case scraper.StepWarmup:
		return 2 // warmup
	case scraper.StepCooldown:
		return 3 // cooldown
	case scraper.StepRest:
		return 1 // rest
	default:
		return 0 // active
	}
}

func generateBaseName(url string) string {
	// Extract last path segment
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	name := parts[len(parts)-1]
	name = strings.ReplaceAll(name, "-", "_")
	// Sanitize
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)

	if name == "" {
		name = "workout"
	}
	return name
}

// expandSteps expands repeats into a flat list for display.
func expandSteps(steps []scraper.WorkoutStep) []scraper.WorkoutStep {
	var expanded []scraper.WorkoutStep

	for i, s := range steps {
		if s.Type == scraper.StepRepeat {
			// The repeat block consists of the N steps immediately before this repeat
			blockSize := s.RepeatSubSteps
			if blockSize <= 0 {
				// Fallback: count non-repeat steps before this one
				blockSize = 0
				for j := i - 1; j >= 0 && steps[j].Type != scraper.StepRepeat; j-- {
					blockSize++
				}
			}
			// Remove the block that was already appended once
			block := make([]scraper.WorkoutStep, blockSize)
			copy(block, expanded[len(expanded)-blockSize:])
			expanded = expanded[:len(expanded)-blockSize]
			// Append it repeated times
			for range s.RepeatCount {
				expanded = append(expanded, block...)
			}
		} else {
			expanded = append(expanded, s)
		}
	}
	return expanded
}

func generateMarkdown(workout *scraper.Workout) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", workout.Name))

	// Build ASCII art
	expanded := expandSteps(workout.Steps)
	asciiArt := buildASCIIArt(expanded)
	sb.WriteString("```\n")
	sb.WriteString(asciiArt)
	sb.WriteString("```\n\n")

	// Write interval table
	sb.WriteString("## Intervalle\n\n")
	sb.WriteString("| # | Phase | Dauer | % FTP |\n")
	sb.WriteString("|---|-------|-------|-------|\n")

	idx := 1
	i := 0
	for i < len(workout.Steps) {
		s := workout.Steps[i]

		if s.Type == scraper.StepRepeat {
			sb.WriteString(fmt.Sprintf("| %d | **Repeat** | — | **%d×** |\n", idx, s.RepeatCount))
			idx++
			i++
			continue
		}

		// Consolidate consecutive ramp steps of same type
		if i+1 < len(workout.Steps) && workout.Steps[i+1].Type == s.Type &&
			s.PowerEndPct == workout.Steps[i+1].PowerStartPct {
			// Find the end of the ramp
			startPct := s.PowerStartPct
			totalDur := s.DurationSec
			endPct := s.PowerEndPct
			j := i + 1
			for j < len(workout.Steps) && workout.Steps[j].Type == s.Type &&
				workout.Steps[j].PowerStartPct == endPct {
				totalDur += workout.Steps[j].DurationSec
				endPct = workout.Steps[j].PowerEndPct
				j++
			}
			phase := stepLabel(s.Type)
			durMin := float64(totalDur) / 60.0
			power := fmt.Sprintf("%d%% → %d%%", startPct, endPct)
			sb.WriteString(fmt.Sprintf("| %d | %s | %.1f min | %s |\n", idx, phase, durMin, power))
			idx++
			i = j
			continue
		}

		phase := stepLabel(s.Type)
		durMin := float64(s.DurationSec) / 60.0
		var power string
		if s.PowerStartPct == s.PowerEndPct {
			power = fmt.Sprintf("%d%%", s.PowerStartPct)
		} else {
			power = fmt.Sprintf("%d%% → %d%%", s.PowerStartPct, s.PowerEndPct)
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %.1f min | %s |\n", idx, phase, durMin, power))
		idx++
		i++
	}

	// Total duration and average power
	totalSec := 0
	weightedPower := 0
	for _, s := range expanded {
		totalSec += s.DurationSec
		avgPct := (s.PowerStartPct + s.PowerEndPct) / 2
		weightedPower += avgPct * s.DurationSec
	}
	avgPower := 0
	if totalSec > 0 {
		avgPower = weightedPower / totalSec
	}
	sb.WriteString(fmt.Sprintf("\n**Gesamtdauer:** %d min | **Ø Leistung:** %d%% FTP\n", totalSec/60, avgPower))

	return sb.String()
}

func stepLabel(t scraper.StepType) string {
	switch t {
	case scraper.StepWarmup:
		return "Warmup"
	case scraper.StepCooldown:
		return "Cooldown"
	case scraper.StepRest:
		return "Rest"
	default:
		return "Active"
	}
}

func buildASCIIArt(steps []scraper.WorkoutStep) string {
	if len(steps) == 0 {
		return ""
	}

	// Chart dimensions
	const height = 12
	const maxWidth = 60

	// Calculate total duration and scale
	totalSec := 0
	for _, s := range steps {
		totalSec += s.DurationSec
	}

	// Build columns: each step gets width proportional to duration
	type column struct {
		pct    int // average % FTP for this step
		width  int // character width
	}

	var cols []column
	for _, s := range steps {
		avgPct := (s.PowerStartPct + s.PowerEndPct) / 2
		w := int(float64(s.DurationSec) / float64(totalSec) * float64(maxWidth))
		if w < 1 {
			w = 1
		}
		cols = append(cols, column{pct: avgPct, width: w})
	}

	// Find max power for scaling
	maxPct := 0
	for _, c := range cols {
		if c.pct > maxPct {
			maxPct = c.pct
		}
	}
	if maxPct == 0 {
		maxPct = 100
	}

	// Build the chart rows (top to bottom)
	totalWidth := 0
	for _, c := range cols {
		totalWidth += c.width
	}

	rows := make([][]byte, height)
	for r := range height {
		rows[r] = make([]byte, totalWidth)
		threshold := maxPct - (maxPct * r / height)
		x := 0
		for _, c := range cols {
			for w := range c.width {
				if c.pct >= threshold {
					rows[r][x+w] = '#'
				} else {
					rows[r][x+w] = ' '
				}
			}
			x += c.width
		}
	}

	// Render
	var sb strings.Builder
	// Y-axis labels
	for r := range height {
		pctLabel := maxPct - (maxPct * r / height)
		sb.WriteString(fmt.Sprintf("%3d%% |", pctLabel))
		sb.Write(rows[r])
		sb.WriteByte('\n')
	}
	// X-axis
	sb.WriteString("     +")
	for range totalWidth {
		sb.WriteByte('-')
	}
	sb.WriteByte('\n')
	// Time label
	sb.WriteString(fmt.Sprintf("      0 min%*s%d min\n", totalWidth-10, "", totalSec/60))

	return sb.String()
}
