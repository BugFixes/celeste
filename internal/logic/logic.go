package logic

import (
	"time"

	"github.com/bugfixes/celeste/internal/config"
	"go.uber.org/zap"
)

type Logic struct {
	Config config.Config
	Logger *zap.SugaredLogger
}

type LogicBug struct {
	LastReported  time.Time
	FirstReported time.Time
	TimesReported int
}

func NewLogic(c config.Config, l *zap.SugaredLogger) *Logic {
	return &Logic{
		Config: c,
		Logger: l,
	}
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
	return time.Since(when) > 1*time.Hour
}

func sinceMoreThanMonth(when time.Time) bool {
	return time.Since(time.Now().AddDate(0, -1, 0)) > time.Since(when)
}

func (l *Logic) ShouldWeReport(lb LogicBug) bool {
	if reportedMoreThanNTimes(lb.TimesReported, 10) {
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
