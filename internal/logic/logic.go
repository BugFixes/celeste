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

func lastReportedLessThanDay(when time.Time) bool {
	return 24*time.Hour > time.Since(when)
}

func lastReportedLessThanWeek(when time.Time) bool {
	return 7*24*time.Hour > time.Since(when)
}

func lastReportedLessThanMonth(when time.Time) bool {
	return time.Since(time.Now().AddDate(0, -1, 0)) > time.Since(when)
}

func lastReportedVeryRecent(when time.Time) bool {
	return 1*time.Hour > time.Since(when)
}

func (l *Logic) ShouldWeReport(lb LogicBug) bool {
	if reportedMoreThanNTimes(10, lb.TimesReported) {
		return true
	}

	if lastReportedVeryRecent(lb.LastReported) {
		return true
	}

	if lastReportedLessThanDay(lb.LastReported) && reportedMoreThanNTimes(5, lb.TimesReported) {
		return true
	}

	if lastReportedLessThanWeek(lb.LastReported) && reportedMoreThanNTimes(5, lb.TimesReported) {
		return true
	}

	if lastReportedLessThanMonth(lb.LastReported) && reportedMoreThanNTimes(5, lb.TimesReported) {
		return true
	}

	return false
}
