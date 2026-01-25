package deltasync

import API "qbit-exp/api"

// MergeServerState merges a DeltaServerState into an existing ServerState.
// Only non-nil fields in delta are applied; nil fields leave existing values unchanged.
func MergeServerState(existing API.ServerState, delta API.DeltaServerState) API.ServerState {
	if delta.AlltimeDl != nil {
		existing.AlltimeDl = *delta.AlltimeDl
	}

	if delta.AlltimeUl != nil {
		existing.AlltimeUl = *delta.AlltimeUl
	}

	if delta.AverageTimeQueue != nil {
		existing.AverageTimeQueue = *delta.AverageTimeQueue
	}

	if delta.ConnectionStatus != nil {
		existing.ConnectionStatus = *delta.ConnectionStatus
	}

	if delta.DHTNodes != nil {
		existing.DHTNodes = *delta.DHTNodes
	}

	if delta.DlInfoData != nil {
		existing.DlInfoData = *delta.DlInfoData
	}

	if delta.DlInfoSpeed != nil {
		existing.DlInfoSpeed = *delta.DlInfoSpeed
	}

	if delta.FreeSpaceOnDisk != nil {
		existing.FreeSpaceOnDisk = *delta.FreeSpaceOnDisk
	}

	if delta.GlobalRatio != nil {
		existing.GlobalRatio = *delta.GlobalRatio
	}

	if delta.QueuedIoJobs != nil {
		existing.QueuedIoJobs = *delta.QueuedIoJobs
	}

	if delta.UpInfoData != nil {
		existing.UpInfoData = *delta.UpInfoData
	}

	if delta.UpInfoSpeed != nil {
		existing.UpInfoSpeed = *delta.UpInfoSpeed
	}

	if delta.TotalBuffersSize != nil {
		existing.TotalBuffersSize = *delta.TotalBuffersSize
	}

	if delta.TotalQueuedSize != nil {
		existing.TotalQueuedSize = *delta.TotalQueuedSize
	}

	if delta.TotalPeerConnections != nil {
		existing.TotalPeerConnections = *delta.TotalPeerConnections
	}

	if delta.TotalWastedSession != nil {
		existing.TotalWastedSession = *delta.TotalWastedSession
	}

	if delta.UseAltSpeedLimits != nil {
		existing.UseAltSpeedLimits = *delta.UseAltSpeedLimits
	}

	return existing
}

// MergeInfo merges a DeltaInfo into an existing Info.
// Only non-nil fields in delta are applied; nil fields leave existing values unchanged.
func MergeInfo(existing API.Info, delta API.DeltaInfo) API.Info {
	if delta.AddedOn != nil {
		existing.AddedOn = *delta.AddedOn
	}

	if delta.AmountLeft != nil {
		existing.AmountLeft = *delta.AmountLeft
	}

	if delta.Category != nil {
		existing.Category = *delta.Category
	}

	if delta.Comment != nil {
		existing.Comment = *delta.Comment
	}

	if delta.CompletionOn != nil {
		existing.CompletionOn = *delta.CompletionOn
	}

	if delta.Dlspeed != nil {
		existing.Dlspeed = *delta.Dlspeed
	}

	if delta.Downloaded != nil {
		existing.Downloaded = *delta.Downloaded
	}

	if delta.DownloadedSession != nil {
		existing.DownloadedSession = *delta.DownloadedSession
	}

	if delta.Eta != nil {
		existing.Eta = *delta.Eta
	}

	if delta.MaxRatio != nil {
		existing.MaxRatio = *delta.MaxRatio
	}

	if delta.Name != nil {
		existing.Name = *delta.Name
	}

	if delta.NumLeechs != nil {
		existing.NumLeechs = *delta.NumLeechs
	}

	if delta.NumSeeds != nil {
		existing.NumSeeds = *delta.NumSeeds
	}

	if delta.Progress != nil {
		existing.Progress = *delta.Progress
	}

	if delta.Ratio != nil {
		existing.Ratio = *delta.Ratio
	}

	if delta.SavePath != nil {
		existing.SavePath = *delta.SavePath
	}

	if delta.Size != nil {
		existing.Size = *delta.Size
	}

	if delta.State != nil {
		existing.State = *delta.State
	}

	if delta.Tags != nil {
		existing.Tags = *delta.Tags
	}

	if delta.Tracker != nil {
		existing.Tracker = *delta.Tracker
	}

	if delta.TimeActive != nil {
		existing.TimeActive = *delta.TimeActive
	}

	if delta.Uploaded != nil {
		existing.Uploaded = *delta.Uploaded
	}

	if delta.UploadedSession != nil {
		existing.UploadedSession = *delta.UploadedSession
	}

	if delta.Upspeed != nil {
		existing.Upspeed = *delta.Upspeed
	}

	return existing
}
