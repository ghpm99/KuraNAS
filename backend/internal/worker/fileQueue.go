package worker

func FileQueueWorker(fileWalkChannel chan FileWalk) {
	fileWalk <- fileWalkChannel
}
