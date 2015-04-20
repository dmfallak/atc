// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/concourse/atc/db"
)

type FakeJobServiceDB struct {
	GetBuildStub        func(buildID int) (db.Build, error)
	getBuildMutex       sync.RWMutex
	getBuildArgsForCall []struct {
		buildID int
	}
	getBuildReturns struct {
		result1 db.Build
		result2 error
	}
	GetJobStub        func(job string) (db.Job, error)
	getJobMutex       sync.RWMutex
	getJobArgsForCall []struct {
		job string
	}
	getJobReturns struct {
		result1 db.Job
		result2 error
	}
	GetRunningBuildsBySerialGroupStub        func(jobName string, serialGroups []string) ([]db.Build, error)
	getRunningBuildsBySerialGroupMutex       sync.RWMutex
	getRunningBuildsBySerialGroupArgsForCall []struct {
		jobName      string
		serialGroups []string
	}
	getRunningBuildsBySerialGroupReturns struct {
		result1 []db.Build
		result2 error
	}
	GetNextPendingBuildBySerialGroupStub        func(jobName string, serialGroups []string) (db.Build, error)
	getNextPendingBuildBySerialGroupMutex       sync.RWMutex
	getNextPendingBuildBySerialGroupArgsForCall []struct {
		jobName      string
		serialGroups []string
	}
	getNextPendingBuildBySerialGroupReturns struct {
		result1 db.Build
		result2 error
	}
}

func (fake *FakeJobServiceDB) GetBuild(buildID int) (db.Build, error) {
	fake.getBuildMutex.Lock()
	fake.getBuildArgsForCall = append(fake.getBuildArgsForCall, struct {
		buildID int
	}{buildID})
	fake.getBuildMutex.Unlock()
	if fake.GetBuildStub != nil {
		return fake.GetBuildStub(buildID)
	} else {
		return fake.getBuildReturns.result1, fake.getBuildReturns.result2
	}
}

func (fake *FakeJobServiceDB) GetBuildCallCount() int {
	fake.getBuildMutex.RLock()
	defer fake.getBuildMutex.RUnlock()
	return len(fake.getBuildArgsForCall)
}

func (fake *FakeJobServiceDB) GetBuildArgsForCall(i int) int {
	fake.getBuildMutex.RLock()
	defer fake.getBuildMutex.RUnlock()
	return fake.getBuildArgsForCall[i].buildID
}

func (fake *FakeJobServiceDB) GetBuildReturns(result1 db.Build, result2 error) {
	fake.GetBuildStub = nil
	fake.getBuildReturns = struct {
		result1 db.Build
		result2 error
	}{result1, result2}
}

func (fake *FakeJobServiceDB) GetJob(job string) (db.Job, error) {
	fake.getJobMutex.Lock()
	fake.getJobArgsForCall = append(fake.getJobArgsForCall, struct {
		job string
	}{job})
	fake.getJobMutex.Unlock()
	if fake.GetJobStub != nil {
		return fake.GetJobStub(job)
	} else {
		return fake.getJobReturns.result1, fake.getJobReturns.result2
	}
}

func (fake *FakeJobServiceDB) GetJobCallCount() int {
	fake.getJobMutex.RLock()
	defer fake.getJobMutex.RUnlock()
	return len(fake.getJobArgsForCall)
}

func (fake *FakeJobServiceDB) GetJobArgsForCall(i int) string {
	fake.getJobMutex.RLock()
	defer fake.getJobMutex.RUnlock()
	return fake.getJobArgsForCall[i].job
}

func (fake *FakeJobServiceDB) GetJobReturns(result1 db.Job, result2 error) {
	fake.GetJobStub = nil
	fake.getJobReturns = struct {
		result1 db.Job
		result2 error
	}{result1, result2}
}

func (fake *FakeJobServiceDB) GetRunningBuildsBySerialGroup(jobName string, serialGroups []string) ([]db.Build, error) {
	fake.getRunningBuildsBySerialGroupMutex.Lock()
	fake.getRunningBuildsBySerialGroupArgsForCall = append(fake.getRunningBuildsBySerialGroupArgsForCall, struct {
		jobName      string
		serialGroups []string
	}{jobName, serialGroups})
	fake.getRunningBuildsBySerialGroupMutex.Unlock()
	if fake.GetRunningBuildsBySerialGroupStub != nil {
		return fake.GetRunningBuildsBySerialGroupStub(jobName, serialGroups)
	} else {
		return fake.getRunningBuildsBySerialGroupReturns.result1, fake.getRunningBuildsBySerialGroupReturns.result2
	}
}

func (fake *FakeJobServiceDB) GetRunningBuildsBySerialGroupCallCount() int {
	fake.getRunningBuildsBySerialGroupMutex.RLock()
	defer fake.getRunningBuildsBySerialGroupMutex.RUnlock()
	return len(fake.getRunningBuildsBySerialGroupArgsForCall)
}

func (fake *FakeJobServiceDB) GetRunningBuildsBySerialGroupArgsForCall(i int) (string, []string) {
	fake.getRunningBuildsBySerialGroupMutex.RLock()
	defer fake.getRunningBuildsBySerialGroupMutex.RUnlock()
	return fake.getRunningBuildsBySerialGroupArgsForCall[i].jobName, fake.getRunningBuildsBySerialGroupArgsForCall[i].serialGroups
}

func (fake *FakeJobServiceDB) GetRunningBuildsBySerialGroupReturns(result1 []db.Build, result2 error) {
	fake.GetRunningBuildsBySerialGroupStub = nil
	fake.getRunningBuildsBySerialGroupReturns = struct {
		result1 []db.Build
		result2 error
	}{result1, result2}
}

func (fake *FakeJobServiceDB) GetNextPendingBuildBySerialGroup(jobName string, serialGroups []string) (db.Build, error) {
	fake.getNextPendingBuildBySerialGroupMutex.Lock()
	fake.getNextPendingBuildBySerialGroupArgsForCall = append(fake.getNextPendingBuildBySerialGroupArgsForCall, struct {
		jobName      string
		serialGroups []string
	}{jobName, serialGroups})
	fake.getNextPendingBuildBySerialGroupMutex.Unlock()
	if fake.GetNextPendingBuildBySerialGroupStub != nil {
		return fake.GetNextPendingBuildBySerialGroupStub(jobName, serialGroups)
	} else {
		return fake.getNextPendingBuildBySerialGroupReturns.result1, fake.getNextPendingBuildBySerialGroupReturns.result2
	}
}

func (fake *FakeJobServiceDB) GetNextPendingBuildBySerialGroupCallCount() int {
	fake.getNextPendingBuildBySerialGroupMutex.RLock()
	defer fake.getNextPendingBuildBySerialGroupMutex.RUnlock()
	return len(fake.getNextPendingBuildBySerialGroupArgsForCall)
}

func (fake *FakeJobServiceDB) GetNextPendingBuildBySerialGroupArgsForCall(i int) (string, []string) {
	fake.getNextPendingBuildBySerialGroupMutex.RLock()
	defer fake.getNextPendingBuildBySerialGroupMutex.RUnlock()
	return fake.getNextPendingBuildBySerialGroupArgsForCall[i].jobName, fake.getNextPendingBuildBySerialGroupArgsForCall[i].serialGroups
}

func (fake *FakeJobServiceDB) GetNextPendingBuildBySerialGroupReturns(result1 db.Build, result2 error) {
	fake.GetNextPendingBuildBySerialGroupStub = nil
	fake.getNextPendingBuildBySerialGroupReturns = struct {
		result1 db.Build
		result2 error
	}{result1, result2}
}

var _ db.JobServiceDB = new(FakeJobServiceDB)