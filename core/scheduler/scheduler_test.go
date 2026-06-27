package scheduler

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/result"
	"github.com/reugn/go-quartz/quartz"
	"github.com/samber/lo"
)

func TestCreateConfigMergeMultipleSources(t *testing.T) {
	cfgMap, err := createConfig([]*Config{
		{JobConfigs: []JobConfig{{Name: "job-a"}}},
		{JobConfigs: []JobConfig{{Name: "job-b"}}},
	})
	if err != nil {
		t.Fatalf("createConfig failed: %v", err)
	}

	if len(cfgMap) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(cfgMap))
	}

	if _, ok := cfgMap["job-a"]; !ok {
		t.Fatalf("job-a config not found")
	}
	if _, ok := cfgMap["job-b"]; !ok {
		t.Fatalf("job-b config not found")
	}
}

func TestCreateConfigSkipNilEntries(t *testing.T) {
	cfgMap, err := createConfig([]*Config{
		nil,
		{JobConfigs: []JobConfig{{Name: "job-c"}}},
	})
	if err != nil {
		t.Fatalf("createConfig failed: %v", err)
	}

	if len(cfgMap) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfgMap))
	}
	if _, ok := cfgMap["job-c"]; !ok {
		t.Fatalf("job-c config not found")
	}
}

func TestCreateConfigDuplicateAcrossSources(t *testing.T) {
	_, err := createConfig([]*Config{
		{JobConfigs: []JobConfig{{Name: "dup-job"}}},
		{JobConfigs: []JobConfig{{Name: "dup-job"}}},
	})
	if err == nil {
		t.Fatalf("expected duplicate config error")
	}
}

func TestInitAndMergeConfigInvalidLocationFallbackUTC(t *testing.T) {
	cfg := initAndMergeConfig("job-loc", &JobConfig{Location: lo.ToPtr("Mars/OlympusMons")})
	if cfg.location != time.UTC {
		t.Fatalf("expected fallback location UTC, got %v", cfg.location)
	}
}

func TestJobConfigToJobDetailOptionsSuspendedFromDisabled(t *testing.T) {
	t.Run("disabled true", func(t *testing.T) {
		cfg := JobConfig{
			Disabled:      lo.ToPtr(true),
			MaxRetries:    lo.ToPtr(2),
			RetryInterval: lo.ToPtr(time.Second),
			Replace:       lo.ToPtr(false),
		}

		opts := cfg.ToJobDetailOptions()
		if !opts.Suspended {
			t.Fatalf("expected Suspended=true when Disabled=true")
		}
	})

	t.Run("disabled false", func(t *testing.T) {
		cfg := JobConfig{
			Disabled:      lo.ToPtr(false),
			MaxRetries:    lo.ToPtr(2),
			RetryInterval: lo.ToPtr(time.Second),
			Replace:       lo.ToPtr(false),
		}

		opts := cfg.ToJobDetailOptions()
		if opts.Suspended {
			t.Fatalf("expected Suspended=false when Disabled=false")
		}
	})
}

func TestTriggerSnapshotConcurrentAccess(t *testing.T) {
	tr := newTrigger(quartz.NewSimpleTrigger(time.Millisecond))

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			_, _ = tr.NextFireTime(time.Now().Add(time.Duration(i) * time.Millisecond).UnixNano())
		}(i)
		go func() {
			defer wg.Done()
			_, _, _ = tr.Snapshot()
		}()
	}
	wg.Wait()
}

func TestJobTaskResultSnapshotConcurrent(t *testing.T) {
	task := &jobTask{
		spec:    &JobSpec{Name: "job-result", Config: initAndMergeConfig("job-result")},
		trigger: newTrigger(quartz.NewSimpleTrigger(time.Millisecond)),
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 200)
	for i := 0; i < 200; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			task.setResult(result.OK([]byte{byte(i % 255)}))
		}(i)
		go func() {
			defer wg.Done()
			job := task.ToJob()
			if job == nil || job.Spec == nil || job.Spec.Name != "job-result" {
				errCh <- errors.New("unexpected job snapshot")
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSchedulerSentinelErrors(t *testing.T) {
	t.Run("create job with empty name", func(t *testing.T) {
		s := &Scheduler{
			log:  log.GetLogger("scheduler-test"),
			jobs: map[string]*jobTask{},
		}

		err := s.createJob(JobSpec{}, nil)
		if !err.IsErr() {
			t.Fatalf("expected create job error")
		}

		if !errors.Is(err.GetErr(), ErrJobNameEmpty) {
			t.Fatalf("expected ErrJobNameEmpty, got: %v", err.GetErr())
		}
	})

	t.Run("create duplicated job", func(t *testing.T) {
		s := &Scheduler{
			log: log.GetLogger("scheduler-test"),
			jobs: map[string]*jobTask{
				"dup-job": {
					spec:   &JobSpec{Name: "dup-job"},
					jobKey: parseJobKey("dup-job"),
					status: StatusRunning,
				},
			},
		}

		err := s.createJob(JobSpec{Name: "dup-job"}, nil)
		if !err.IsErr() {
			t.Fatalf("expected create duplicated job error")
		}

		if !errors.Is(err.GetErr(), ErrJobAlreadyExists) {
			t.Fatalf("expected ErrJobAlreadyExists, got: %v", err.GetErr())
		}
	})

	t.Run("get missing job", func(t *testing.T) {
		s := &Scheduler{jobs: map[string]*jobTask{}}

		res := s.getJob("missing-job")
		if !res.IsErr() {
			t.Fatalf("expected get missing job error")
		}

		if !errors.Is(res.GetErr(), ErrJobNotFound) {
			t.Fatalf("expected ErrJobNotFound, got: %v", res.GetErr())
		}
	})
}
