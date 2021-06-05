package backend

/*
	The scheduler is initialized at start time with maintenance tasks and subscriptions
	It maintains the list of tasks to be performed

	At regular time, the task list is checked. Tasks not yet ran are launched.

*/
