package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type StepType int

const (
	StepWarmup StepType = iota
	StepActive
	StepRest
	StepCooldown
	StepRepeat
)

type WorkoutStep struct {
	Type           StepType
	DurationSec    int
	PowerStartPct  int // % FTP start (for ramps)
	PowerEndPct    int // % FTP end (for ramps, or same as start for steady)
	RepeatFrom     int // for repeat steps: index of first step in block
	RepeatCount    int // for repeat steps: number of repetitions
	RepeatSubSteps int // for repeat steps: how many preceding steps form the block
}

type Workout struct {
	Name  string
	Steps []WorkoutStep
}

func FetchWorkout(url string) (*Workout, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	workout := &Workout{}

	// Extract workout name from h1
	doc.Find("h1").First().Each(func(i int, s *goquery.Selection) {
		workout.Name = strings.TrimSpace(s.Text())
	})

	if workout.Name == "" {
		workout.Name = "Workout"
	}

	// Extract workout description from textbar divs
	var descriptions []string
	doc.Find("div.textbar").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			descriptions = append(descriptions, text)
		}
	})

	if len(descriptions) == 0 {
		return nil, fmt.Errorf("no workout steps found on page")
	}

	// Parse each description line into workout steps
	for i, desc := range descriptions {
		steps, err := parseDescription(desc, i == 0, i == len(descriptions)-1)
		if err != nil {
			return nil, fmt.Errorf("failed to parse step %d (%q): %w", i, desc, err)
		}
		workout.Steps = append(workout.Steps, steps...)
	}

	return workout, nil
}

// parseDescription parses a workout description text like:
// "10min from 30 to 80% FTP"
// "5x 5min @ 110% FTP, 5min @ 55% FTP"
// "5min @ 75% FTP"
// "15min @ 65rpm, from 65 to 92% FTP"
// "30sec @ 109% FTP"
func parseDescription(desc string, isFirst, isLast bool) ([]WorkoutStep, error) {
	desc = normalizeText(desc)

	// Strip cadence annotations like "@ 65rpm," or "@ 85rpm,"
	cadenceRe := regexp.MustCompile(`@\s*\d+\s*rpm\s*,?\s*`)
	desc = cadenceRe.ReplaceAllString(desc, "")
	desc = normalizeText(desc)

	// Pattern: ramp "Xmin from A to B% FTP"
	rampRe := regexp.MustCompile(`(\d+)\s*min\s+from\s+(\d+(?:\.\d+)?)\s*(?:%\s*FTP\s*)?\s*to\s+(\d+(?:\.\d+)?)\s*%\s*FTP`)
	if m := rampRe.FindStringSubmatch(desc); m != nil {
		dur, _ := strconv.Atoi(m[1])
		startPct := parsePercent(m[2])
		endPct := parsePercent(m[3])

		stepType := StepActive
		if isFirst {
			stepType = StepWarmup
		} else if isLast {
			stepType = StepCooldown
		}

		return buildRampSteps(dur*60, startPct, endPct, stepType), nil
	}

	// Pattern: repeated intervals "Nx Xmin @ A% FTP, Xmin @ B% FTP"
	repeatRe := regexp.MustCompile(`(\d+)x\s+(.+)`)
	if m := repeatRe.FindStringSubmatch(desc); m != nil {
		repeatCount, _ := strconv.Atoi(m[1])
		intervalsText := m[2]

		// Parse sub-intervals separated by comma
		parts := strings.Split(intervalsText, ",")
		var subSteps []WorkoutStep
		for _, part := range parts {
			step, err := parseSingleInterval(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			subSteps = append(subSteps, step)
		}

		// Build result: sub-steps + repeat step
		var result []WorkoutStep
		repeatFromIdx := -1 // will be set by caller based on total step count
		result = append(result, subSteps...)
		result = append(result, WorkoutStep{
			Type:           StepRepeat,
			RepeatFrom:     repeatFromIdx,
			RepeatCount:    repeatCount,
			RepeatSubSteps: len(subSteps),
		})
		return result, nil
	}

	// Pattern: single steady interval "Xmin @ A% FTP"
	step, err := parseSingleInterval(desc)
	if err != nil {
		return nil, err
	}

	if isFirst {
		step.Type = StepWarmup
	} else if isLast {
		step.Type = StepCooldown
	}

	return []WorkoutStep{step}, nil
}

// parseSingleInterval parses "5min @ 110% FTP", "30sec @ 109% FTP", "30s @ 55% FTP"
func parseSingleInterval(text string) (WorkoutStep, error) {
	// Handle "Xmin @ Y% FTP" (supports decimal %)
	re := regexp.MustCompile(`(\d+)\s*min\s*@\s*(\d+(?:\.\d+)?)\s*%\s*FTP`)
	if m := re.FindStringSubmatch(text); m != nil {
		dur, _ := strconv.Atoi(m[1])
		pct := parsePercent(m[2])

		stepType := StepActive
		if pct < 70 {
			stepType = StepRest
		}

		return WorkoutStep{
			Type:          stepType,
			DurationSec:   dur * 60,
			PowerStartPct: pct,
			PowerEndPct:   pct,
		}, nil
	}

	// Handle "Xs @ Y% FTP" or "Xsec @ Y% FTP" (seconds, supports decimal %)
	reSec := regexp.MustCompile(`(\d+)\s*(?:sec|s)\s*@\s*(\d+(?:\.\d+)?)\s*%\s*FTP`)
	if m := reSec.FindStringSubmatch(text); m != nil {
		dur, _ := strconv.Atoi(m[1])
		pct := parsePercent(m[2])

		stepType := StepActive
		if pct < 70 {
			stepType = StepRest
		}

		return WorkoutStep{
			Type:          stepType,
			DurationSec:   dur,
			PowerStartPct: pct,
			PowerEndPct:   pct,
		}, nil
	}

	return WorkoutStep{}, fmt.Errorf("could not parse interval: %q", text)
}

// buildRampSteps creates multiple 1-minute steps for a ramp
func buildRampSteps(totalSec, startPct, endPct int, stepType StepType) []WorkoutStep {
	numSteps := totalSec / 60
	if numSteps < 1 {
		numSteps = 1
	}

	steps := make([]WorkoutStep, numSteps)
	for i := range numSteps {
		pctStart := startPct + (endPct-startPct)*i/numSteps
		pctEnd := startPct + (endPct-startPct)*(i+1)/numSteps

		steps[i] = WorkoutStep{
			Type:          stepType,
			DurationSec:   totalSec / numSteps,
			PowerStartPct: pctStart,
			PowerEndPct:   pctEnd,
		}
	}
	return steps
}

func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	// Collapse multiple spaces
	spaceRe := regexp.MustCompile(`\s+`)
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// parsePercent parses a percent string that may be decimal (e.g. "109.44999") and rounds to int
func parsePercent(s string) int {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Fallback to integer parsing
		v, _ := strconv.Atoi(s)
		return v
	}
	return int(f + 0.5)
}
