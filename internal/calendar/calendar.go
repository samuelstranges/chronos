package calendar

import (
	"strconv"
	"time"
)

type Calendar struct {
	CurrentDay  *Day
	CurrentWeek *Week
}

func NewCalendar(currentDay *Day) *Calendar {
	c := &Calendar{CurrentDay: currentDay}

	c.CurrentWeek = NewWeek()
	c.UpdateWeek()

	return c
}

func (c *Calendar) setWeekLimits() {
	c.RoundTime()
	d := c.CurrentDay.Date

	diffToSunday := d.Weekday()
	diffToSaturday := 6 - d.Weekday()

	c.CurrentWeek.StartDate = d.AddDate(0, 0, -int(diffToSunday))
	c.CurrentWeek.EndDate = d.AddDate(0, 0, int(diffToSaturday))
}

func (c *Calendar) FormatWeekBody() string {
	startDay := c.CurrentWeek.StartDate
	endDay := c.CurrentWeek.EndDate
	month := endDay.Month().String()

	return month + " " + strconv.Itoa(startDay.Day()) + " to " + strconv.Itoa(endDay.Day())
}

func (c *Calendar) UpdateWeek() {
	c.setWeekLimits()

	for i := range c.CurrentWeek.Days {
		d := c.CurrentWeek.StartDate.AddDate(0, 0, i)
		c.CurrentWeek.Days[i].Date = d
	}
}

func (c *Calendar) RoundTime() {
	min := c.CurrentDay.Date.Minute()

	if min >= 0 && min <= 14 {
		c.CurrentDay.Date = c.CurrentDay.Date.Add(time.Minute * time.Duration(-min))
	} else if min >= 14 && min <= 44 {
		diff := 30 - min
		c.CurrentDay.Date = c.CurrentDay.Date.Add(time.Minute * time.Duration(diff))
	} else {
		diff := 60 - min
		c.CurrentDay.Date = c.CurrentDay.Date.Add(time.Minute * time.Duration(diff))
	}
}

func (c *Calendar) JumpToToday() {
	now := time.Now()
	c.CurrentDay.Date = time.Date(now.Year(), now.Month(), now.Day(), c.CurrentDay.Date.Hour(), c.CurrentDay.Date.Minute(), 0, 0, now.Location())
	c.UpdateWeek()
}

func (c *Calendar) UpdateToNextWeek() {
	c.CurrentDay.Date = c.CurrentDay.Date.AddDate(0, 0, 7)
	c.UpdateWeek()
}

func (c *Calendar) UpdateToPrevWeek() {
	c.CurrentDay.Date = c.CurrentDay.Date.AddDate(0, 0, -7)
	c.UpdateWeek()
}

func (c *Calendar) UpdateToNextDay() {
	c.CurrentDay.Date = c.CurrentDay.Date.AddDate(0, 0, 1)
	c.UpdateWeek()
}

func (c *Calendar) UpdateToPrevDay() {
	c.CurrentDay.Date = c.CurrentDay.Date.AddDate(0, 0, -1)
	c.UpdateWeek()
}

func (c *Calendar) UpdateToNextTime() {
	c.CurrentDay.Date = c.CurrentDay.Date.Add(time.Minute * time.Duration(30))
	c.UpdateWeek()
}

func (c *Calendar) UpdateToPrevTime() {
	c.CurrentDay.Date = c.CurrentDay.Date.Add(time.Minute * time.Duration(-30))
	c.UpdateWeek()
}

func (c *Calendar) GotoTime(hour, minute int) {
	currentDate := c.CurrentDay.Date
	c.CurrentDay.Date = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), hour, minute, 0, 0, currentDate.Location())
	c.UpdateWeek()
}

func (c *Calendar) GotoDate(year, month, day int) {
	currentDate := c.CurrentDay.Date
	c.CurrentDay.Date = time.Date(year, time.Month(month), day, currentDate.Hour(), currentDate.Minute(), 0, 0, currentDate.Location())
	c.UpdateWeek()
}

func (c *Calendar) GetDayFromTime(time time.Time) *Day {
	for _, v := range c.CurrentWeek.Days {
		vYear, vMonth, vDay := v.Date.Date()
		tYear, tMonth, tDay := time.Date()
		if vYear == tYear && vMonth == tMonth && vDay == tDay {
			return v
		}
	}
	return &Day{}
}

func (c *Calendar) UpdateToNextMonth() {
	// Go to the first day of the next month
	currentDate := c.CurrentDay.Date
	nextMonth := currentDate.AddDate(0, 1, 0)
	c.CurrentDay.Date = time.Date(nextMonth.Year(), nextMonth.Month(), 1, currentDate.Hour(), currentDate.Minute(), 0, 0, currentDate.Location())
	c.UpdateWeek()
}

func (c *Calendar) UpdateToPrevMonth() {
	// Go to the first day of the previous month
	currentDate := c.CurrentDay.Date
	prevMonth := currentDate.AddDate(0, -1, 0)
	c.CurrentDay.Date = time.Date(prevMonth.Year(), prevMonth.Month(), 1, currentDate.Hour(), currentDate.Minute(), 0, 0, currentDate.Location())
	c.UpdateWeek()
}
