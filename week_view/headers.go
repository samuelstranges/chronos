package week_view

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/week_view_shared"
)

func getMonthYearHeader(weekStart time.Time) string {
	weekEnd := weekStart.AddDate(0, 0, types.MaxDayIndex)

	if weekStart.Year() == weekEnd.Year() && weekStart.Month() == weekEnd.Month() {
		return weekStart.Format("January 2006")
	} else if weekStart.Year() == weekEnd.Year() {
		return fmt.Sprintf("%s - %s",
			weekStart.Format("Jan"),
			weekEnd.Format("Jan 2006"))
	} else {
		return fmt.Sprintf("%s - %s",
			weekStart.Format("Jan 2006"),
			weekEnd.Format("Jan 2006"))
	}
}

func renderMonthYearHeader(m types.WeekModel) string {
	monthYearText := getMonthYearHeader(m.CurrentlyViewedWeek)
	return week_view_shared.MonthYearHeaderStyle.Width(m.Width).Render(monthYearText)
}

func getDayHeader(m types.WeekModel, weekStartDate time.Time, daysFromFirstDayOfWeek int) string {
	date := weekStartDate.AddDate(0, 0, daysFromFirstDayOfWeek)
	dayStr := date.Format("Mon 2") // "Mon 2" gives "Mon 15", "Tue 16", etc.
	dayHeaderStyle := week_view_shared.GenericHeaderStyle.Width(m.CachedCellWidth)
	return dayHeaderStyle.Render(dayStr)
}

func renderDaysHeadersHorizontally(m types.WeekModel, weekStartDate time.Time) string {
	cells := make([]string, 0, types.DaysPerWeek+1)
	cells = append(cells, week_view_shared.TimeStyleHeader.Render(""))

	for dayOfWeek := range types.DaysPerWeek {
		cells = append(cells, getDayHeader(m, weekStartDate, dayOfWeek))
	}

	return lipgloss.NewStyle().
		Width(m.Width).
		Background(lipgloss.Color(types.HeaderBackgroundColor)).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
}
