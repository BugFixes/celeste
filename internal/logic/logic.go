package logic

import (
	"time"

	"github.com/bugfixes/celeste/internal/config"
)

type Logic struct {
	Config config.Config
}

type LogicBug struct {
	LastReported  time.Time
	FirstReported time.Time
	TimesReported int
}

func NewLogic(c config.Config) *Logic {
	return &Logic{
		Config: c,
	}
}

func firstReport(when time.Time) bool {
	return when == time.Now()
}

func reportedMoreThanNTimes(times, n int) bool {
	return times > n
}

func sinceLessThanDay(when time.Time) bool {
	return time.Since(when) > 24*time.Hour
}

func sinceLessThanWeek(when time.Time) bool {
	return time.Since(when) > 7*24*time.Hour
}

func sinceLessThanMonth(when time.Time) bool {
	return time.Since(when) > time.Since(time.Now().AddDate(0, -1, 0))
}

func sinceLessThanHour(when time.Time) bool {
	return 1*time.Hour > time.Since(when)
}

func sinceMoreThanMonth(when time.Time) bool {
	return time.Since(time.Now().AddDate(0, -1, 0)) > time.Since(when)
}

// ShouldWeReport
// nolint: gocyclo
func (l *Logic) ShouldWeReport(lb LogicBug) bool {
	// no need to report everything is hosted locally
	if l.Config.KeepLocal {
		return false
	}

	if reportedMoreThanNTimes(lb.TimesReported, 10) {
		return true
	}

	if firstReport(lb.FirstReported) {
		return true
	}

	if sinceLessThanHour(lb.LastReported) {
		return true
	}

	if sinceLessThanDay(lb.LastReported) && reportedMoreThanNTimes(lb.TimesReported, 5) {
		return true
	}

	if sinceLessThanWeek(lb.LastReported) && reportedMoreThanNTimes(lb.TimesReported, 5) {
		return true
	}

	if sinceLessThanMonth(lb.LastReported) && reportedMoreThanNTimes(5, lb.TimesReported) {
		return true
	}

	if sinceMoreThanMonth(lb.FirstReported) {
		return true
	}

	return false
}
