package process_result

var (
	SetElasticInfo = setElasticInfo
	SetWebexInfo   = setWebexInfo
	SetSlackInfo   = setSlackInfo
	SetJiraInfo    = setJiraInfo

	PrepareMessage = prepareMessage
)

type (
	Info = info
)

func (i *info) SetRundId(runId int64) {
	i.runID = runId
}
