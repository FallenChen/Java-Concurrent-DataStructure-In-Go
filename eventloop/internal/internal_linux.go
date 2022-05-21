
package internal

type Poll struct {
	fd	int	// epoll fd
	wfd 	int	// wake  fd
	notes 	noteQueue
}