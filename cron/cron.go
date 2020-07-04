package cron

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"errors"
)

var loc = time.Local

func ChangeLoc(newLocation *time.Location) {
	loc = newLocation
}

const MAXJOBNUM = 10000

type Job struct {
	interval uint64  // pause interval * unit bettween runs
	jobFunc  string // the job jobFunc to run, func[jobFunc]
	unit     string // time units, e.g. 'minute', 'hour'....
	atTime   string // optional time at which this job runs

	lastRun  time.Time                  // datetime of last run
	nextRun  time.Time                  // datetime of next run
	period   time.Duration                  // cache the period between last an next run
	startDay time.Weekday               // specific day of the week to start on
	funcs    map[string]interface{}     // map for the function task store
	fparams  map[string]([]interface{}) // map for function and params of function
	jobid    string
	timeZone string
}

func NewJob(interval uint64, id string, timezone string) *Job {
	return &Job{
		interval: interval,
		jobFunc: "",
		unit: "",
		atTime: "",
		lastRun: time.Unix(0, 0),
		nextRun: time.Unix(0, 0),
		period: 0,
		startDay: time.Sunday,
		funcs: make(map[string]interface{}),
		fparams: make(map[string]([]interface{})),
		jobid: id,
		timeZone: timezone,
	}
}

// True if the job should be run now
func (j *Job) shouldRun() bool {
	return time.Now().After(j.nextRun)
}

// Run the job and immediately reschedule it
func (j *Job) run() (result []reflect.Value, err error) {
	f := reflect.ValueOf(j.funcs[j.jobFunc])
	params := j.fparams[j.jobFunc]
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of param is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	j.lastRun = ReWriteLastTime(time.Now(), j.timeZone)
	j.scheduleNextRun()
	return
}

// for given function fn, get the name of function.
func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// specifies the jobFunc that should be called every time the job runs
func (j *Job) Do(jobFunc interface{}, params ...interface{}) {
	typ := reflect.TypeOf(jobFunc)
	if typ.Kind() != reflect.Func {
		panic("only function can be schedule into the job queue.")
	}
	fname := getFunctionName(jobFunc)
	j.funcs[fname] = jobFunc
	j.fparams[fname] = params
	j.jobFunc = fname

	j.scheduleNextRun()
}

func formatTime(t string) (hour, min int, err error) {
	var er = errors.New("time format error")
	ts := strings.Split(t, ":")
	if len(ts) != 2 {
		err = er
		return
	}
	hour, err = strconv.Atoi(ts[0])
	if err != nil {
		return
	}
	min, err = strconv.Atoi(ts[1])
	if err != nil {
		return
	}
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		err = er
		return
	}
	return hour, min, nil
}

// s.Every(1).Day().At("10:30").Do(task)
func (j *Job) At(t string) *Job {
	hour, min, err := formatTime(t)
	if err != nil {
		panic(err)
	}
	mock := time.Date(time.Now().Year(),  time.Now().Month(), time.Now().Day(), int(hour), int(min), 0, 0, loc)
	if j.unit == "days" {
		if time.Now().After(mock) {
			j.lastRun = ReWriteLastTime(mock, j.timeZone)
			fmt.Println("[175] Last Run Changed with value ", j.lastRun)
		} else {
			j.lastRun = ReWriteLastTime(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, hour, min, 0, 0, loc), j.timeZone)
			fmt.Println("[175] Last Run Changed with value ", j.lastRun)
		}
	} else if j.unit == "weeks" {
		if j.startDay != time.Now().Weekday() || (time.Now().After(mock) && j.startDay == time.Now().Weekday()) {
			i := mock.Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = ReWriteLastTime(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), hour, min, 0, 0, loc), j.timeZone)
			fmt.Println("[187] Last Run Changed with value ", j.lastRun)
		} else {
			j.lastRun = ReWriteLastTime(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-7, hour, min, 0, 0, loc), j.timeZone)
			fmt.Println("[190] Last Run Changed with value ", j.lastRun)
		}
	}
	return j
}

func (j *Job) scheduleNextRun() {
	fmt.Println("scheduleNextRun")
	if j.lastRun == time.Unix(0, 0) {
		if j.unit == "weeks" {
			i := time.Now().Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = ReWriteLastTime(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), 0, 0, 0, 0, loc), j.timeZone)
			fmt.Println("[207] Last Run Changed with value ", j.lastRun)
		} else {
			j.lastRun = ReWriteLastTime(time.Now(), j.timeZone)
			fmt.Println("[209] Last Run Changed with value ", j.lastRun)
		}
	}

	if j.period != 0 {
		if j.unit == "days" || j.unit == "weeks" {
			j.nextRun = ReWriteTime(j.lastRun.Add(j.period * time.Second), j.timeZone, j.lastRun)
			fmt.Println("[218] Next Run Changed with value ", j.lastRun)
		} else {
			j.nextRun = j.lastRun.Add(j.period * time.Second)
			fmt.Println("[221] Next Run Changed with value ", j.lastRun)
		}
	} else {
		switch j.unit {
		case "minutes":
			j.period = time.Duration(j.interval * 60)
			break
		case "hours":
			j.period = time.Duration(j.interval * 60 * 60)
			break
		case "days":
			j.period = time.Duration(j.interval * 60 * 60 * 24)
			break
		case "weeks":
			j.period = time.Duration(j.interval * 60 * 60 * 24 * 7)
			break
		case "seconds":
			j.period = time.Duration(j.interval)
		}
		if j.unit == "days" || j.unit == "weeks"{
			j.nextRun= ReWriteTime(j.lastRun.Add(j.period * time.Second), j.timeZone, j.lastRun)
			fmt.Println("[243] Next Run Changed with value ", j.lastRun)
		} else {
			j.nextRun = j.lastRun.Add(j.period * time.Second)
			fmt.Println("[245] Next Run Changed with value ", j.lastRun)
		}
	}
}

func ReWriteTime(nextRun time.Time, timeZone string, lastRun time.Time) time.Time {
	var constTimeZone = []string{"AST", "WAT", "CAT", "IST"}
	for _, val := range constTimeZone {
		if val == timeZone {
			return nextRun
		}
	}

	_, month, _ := nextRun.Date()
	if int(month) >= 3 && int(month) <= 11 {
		return nextRun.Add(time.Duration(-60) * time.Minute)
	}
	return nextRun
}

func ReWriteLastTime(lastRun time.Time, timeZone string) time.Time {
	var constTimeZone = []string{"AST", "WAT", "CAT", "IST"}
	for _, val := range constTimeZone {
		if val == timeZone {
			return lastRun
		}
	}

	_, month, _ := lastRun.Date()
	if int(month) >= 3 && int(month) < 11 {
		return lastRun.Add(time.Duration(60) * time.Minute)
	}
	return lastRun
}

// NextScheduledTime returns the time of when this job is to run next
func (j *Job) NextScheduledTime() time.Time {
	return j.nextRun
}

// Set the unit with second
func (j *Job) Second() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Seconds()
	return
}

// Set the unit with seconds
func (j *Job) Seconds() (job *Job) {
	j.unit = "seconds"
	return j
}

// Set the unit  with minute, which interval is 1
func (j *Job) Minute() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Minutes()
	return
}

//set the unit with minute
func (j *Job) Minutes() (job *Job) {
	j.unit = "minutes"
	return j
}

//set the unit with hour, which interval is 1
func (j *Job) Hour() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Hours()
	return
}

// Set the unit with hours
func (j *Job) Hours() (job *Job) {
	j.unit = "hours"
	return j
}

// Set the job's unit with day, which interval is 1
func (j *Job) Day() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Days()
	return
}

// Set the job's unit with days
func (j *Job) Days() *Job {
	j.unit = "days"
	return j
}

// s.Every(1).Monday().Do(task)
// Set the start day with Monday
func (j *Job) Monday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 1
	job = j.Weeks()
	return
}

// Set the start day with Tuesday
func (j *Job) Tuesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 2
	job = j.Weeks()
	return
}

// Set the start day woth Wednesday
func (j *Job) Wednesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 3
	job = j.Weeks()
	return
}

// Set the start day with thursday
func (j *Job) Thursday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 4
	job = j.Weeks()
	return
}

// Set the start day with friday
func (j *Job) Friday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 5
	job = j.Weeks()
	return
}

// Set the start day with saturday
func (j *Job) Saturday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 6
	job = j.Weeks()
	return
}

// Set the start day with sunday
func (j *Job) Sunday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 0
	job = j.Weeks()
	return
}

//Set the units as weeks
func (j *Job) Weeks() *Job {
	j.unit = "weeks"
	return j
}


type Scheduler struct {
	jobs [MAXJOBNUM]*Job
	size int // size of jobs which jobs holding
}

// implements the sort.Interface{}

func (s *Scheduler) Len() int {
	return s.size
}

func (s *Scheduler) Less(i, j int) bool {
	return s.jobs[j].nextRun.After(s.jobs[i].nextRun) // 比较两个任务的下次执行时间先后
}

func (s *Scheduler) Swap(i, j int) {
	s.jobs[i], s.jobs[j] = s.jobs[j], s.jobs[i]
}

// Get the current runnable jobs, which should run is true
func (s *Scheduler) getRunnableJobs() (running_jobs [MAXJOBNUM]*Job, n int) {
	runnable_jobs := [MAXJOBNUM]*Job{}
	n = 0
	sort.Sort(s)
	for i := 0; i < s.size; i++ {
		if s.jobs[i].shouldRun() {
			runnable_jobs[n] = s.jobs[i]
			n++
		} else {
			break
		}
	}
	return runnable_jobs, n
}

// datetime when the next job should run
func (s *Scheduler) NextRun() (*Job, time.Time) {
	if s.size <= 0 {
		return nil, time.Now()
	}
	sort.Sort(s)
	return s.jobs[0], s.jobs[0].nextRun
}

// schedule a new periodic job
func (s *Scheduler) Every(interval uint64, id string, timeZone string) *Job {
	job := NewJob(interval, id, timeZone)
	s.jobs[s.size] = job
	s.size++
	return job
}

// run all the jobs that are scheduled to run
func (s *Scheduler) RunPending() {
	runnableJobs, n := s.getRunnableJobs()
	if n != 0 {
		for i := 0; i < n; i++ {
			runnableJobs[i].run()
		}
	}
}

// run all the jobs regardless if they are sccheduled to run or not
func (s *Scheduler) RunAll() {
	for i := 0; i < s.size; i++ {
		s.jobs[i].run()
	}
}

// run all the jobs with delay seconds
func (s *Scheduler) RunAllWithDelay(d int) {
	for i := 0; i < s.size; i++ {
		s.jobs[i].run()
		time.Sleep(time.Duration(d))
	}
}

// remove specific job j
func (s *Scheduler) Remove(j interface{}, jobid string) bool {
	i := 0
	res := false
	for ; i < s.size; i++ {
		if s.jobs[i].jobFunc == getFunctionName(j) && s.jobs[i].jobid == jobid {
			res = true
			break
		}
	}
	for j := (i + 1); j < s.size; j++ {
		s.jobs[i] = s.jobs[j]
		i++
	}
	return res
}

// remove all scheduled jobs
func (s *Scheduler) Clear() {
	for i := 0; i < s.size; i++ {
		s.jobs[i] = nil
	}
	s.size = 0
}

// start all the pending jobs
// add seconds ticker
func (s *Scheduler) Start() chan bool {
	stopped := make(chan bool, 1)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.RunPending()
			case <-stopped:
				return
			}
		}
	}()
	return stopped
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs: [MAXJOBNUM]*Job{},
		size: 0,
	}
}


// The following methods are shortcuts for not having to
// create a scheduler instance

var defaultScheduler = NewScheduler()
var jobs = defaultScheduler.jobs

// schedule a new periodic job
func Every(interval uint64, id string, timeZone string) *Job {
	return defaultScheduler.Every(interval, id, timeZone)
}

func RunPending() {
	defaultScheduler.RunPending()
}

func RunAll() {
	defaultScheduler.RunAll()
}

func RunWithDelay(d int) {
	defaultScheduler.RunAllWithDelay(d)
}

func Start() chan bool {
	return defaultScheduler.Start()
}

func Clear() {
	defaultScheduler.Clear()
}

func Remove(j interface{}, jobid string) {
	defaultScheduler.Remove(j, jobid)
}

func NextRun() (job *Job, time time.Time) {
	return defaultScheduler.NextRun()
}



















