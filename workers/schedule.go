package workers

import (
	"fmt"
	"time"
)

//hourPeriod для описания промежутков времени
type HourPeriod struct {
	HourFrom *uint8
	HourTo   *uint8
}

//days - временные промежутки в контексте дней
type Days map[time.Weekday]*HourPeriod

//schedule для описания расписания
type Schedule struct {
	//days описывает временные рамки
	Days Days

	//delaySeconds задержка между исполнением
	DelaySeconds uint64

	//isActive признак активности
	IsActive bool

	//Текущая версия расписания
	Version int

	//Время простоя в случае проблемы
	FailureTimeOut time.Duration
}

func isHourValid(hour *uint8) bool {
	if hour == nil {
		return true
	}

	if *hour < 0 || *hour > 23 {
		return false
	}

	return true
}

func InitDays() Days {
	return map[time.Weekday]*HourPeriod{}
}

func (d Days) AddPeriod(wd time.Weekday, hourFrom *uint8, hourTo *uint8) error {
	hp, err := createHourPeriod(hourFrom, hourTo)
	if err != nil {
		return err
	}

	d[wd] = hp

	return nil
}

func createHourPeriod(hourFrom *uint8, hourTo *uint8) (*HourPeriod, error) {
	if !isHourValid(hourFrom) {
		return nil, fmt.Errorf("Не валиден параметр веремени (h_from): %d", hourFrom)
	}

	if !isHourValid(hourTo) {
		return nil, fmt.Errorf("Не валиден параметр веремени (h_to): %d", hourTo)
	}

	return &HourPeriod{
		HourFrom: hourFrom,
		HourTo:   hourTo,
	}, nil
}

func CreateSchedule(d Days) *Schedule {
	return &Schedule{
		Days: d,
	}
}

func EveryDay() *Schedule {
	return &Schedule{
		Days: map[time.Weekday]*HourPeriod{
			time.Sunday:    &HourPeriod{},
			time.Monday:    &HourPeriod{},
			time.Tuesday:   &HourPeriod{},
			time.Wednesday: &HourPeriod{},
			time.Thursday:  &HourPeriod{},
			time.Friday:    &HourPeriod{},
			time.Saturday:  &HourPeriod{},
		},
	}
}

func (s *Schedule) Since(hourFrom *uint8) (*Schedule, error) {
	if !isHourValid(hourFrom) {
		return nil, fmt.Errorf("Не валиден параметр веремени (h_from): %d", hourFrom)
	}

	for key := range s.Days {
		val := s.Days[key]

		val.HourFrom = hourFrom

		s.Days[key] = val
	}

	return s, nil
}

func (s *Schedule) To(hourTo *uint8) (*Schedule, error) {
	if !isHourValid(hourTo) {
		return nil, fmt.Errorf("Не валиден параметр веремени (h_to): %d", hourTo)
	}

	for key := range s.Days {
		val := s.Days[key]

		val.HourTo = hourTo

		s.Days[key] = val
	}

	return s, nil
}

func (s *Schedule) WithDelay(sec uint64) *Schedule {
	s.DelaySeconds = sec

	return s
}

func (s *Schedule) WithFailureTimeout(d time.Duration) *Schedule {
	s.FailureTimeOut = d

	return s
}

func (s *Schedule) SetVersion(ver int) *Schedule {
	s.Version = ver

	return s
}

func (s *Schedule) SetIsActive(sign bool) *Schedule {
	s.IsActive = sign

	return s
}

func datediff(dateFrom time.Time, dateTo time.Time) time.Duration {
	return time.Until(dateTo) - time.Until(dateFrom)
}

//createDateFromSchedule создает временной диапазон работы воркеров
func createDateFromSchedule(baseDate time.Time, hourFrom *uint8, hourTo *uint8) (dateFrom time.Time, dateTo time.Time) {
	if hourFrom == nil {
		dateFrom = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		dateFrom = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), int(*hourFrom), 0, 0, 0, time.UTC)
	}

	if hourTo == nil {
		dateTo = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	} else {
		dateTo = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), int(*hourTo), 0, 0, 0, time.UTC)
	}

	return dateFrom, dateTo
}

//calculateDelay расчитывает задержку до запуска воркера, либо
//в случае невозможности, возвращает статус неактивности воркера
func (s *Schedule) calculateDelay() (wait time.Duration, isActive bool) {
	now := time.Now().UTC()
	cWd := now.Weekday()

	//Кол-во дней до старта воркера
	var countDays int = 8

	var wd time.Weekday

	hourPeriod, ok := s.Days[cWd]
	if ok {
		dateFrom, dateTo := createDateFromSchedule(now, hourPeriod.HourFrom, hourPeriod.HourTo)

		timeToStart := datediff(now, dateFrom)
		if timeToStart > 0 {
			return timeToStart, true
		}

		if timeToStart <= 0 && datediff(now, dateTo) > 0 {
			return 0, true
		}
	}

	for key := range s.Days {
		if key == cWd {
			continue
		}

		if key <= cWd {
			key += 7
		}

		curDiff := int(key - cWd)

		if curDiff < countDays {
			countDays = curDiff
			wd = key
		}
	}

	if countDays == 8 {
		return 0, false
	}

	dFrom, _ := createDateFromSchedule(now.AddDate(0, 0, countDays), s.Days[wd].HourFrom, nil)

	return datediff(now, dFrom), true
}
