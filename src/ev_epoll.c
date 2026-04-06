#include <errno.h>
#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/epoll.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/types.h>
#include <unistd.h>

#define MAXBUF 1000
#define MAXEVENTS 20

void
usage_error(const char *prog_name)
{
    fprintf(stderr, "Usage: %s filename\n", prog_name);
    exit(1);
}

void
err_exit(char *error)
{
    perror(error);
    exit(1);
}

int
main(int argc, char *argv[])
{
    setbuf(stdout, NULL);

    if (argc < 2)
        usage_error(argv[0]);

    int epfd = epoll_create1(0);
    if (epfd == -1)
        err_exit("epoll_create");

    int num_open_fds = 0;

    struct epoll_event ev;
    for (int i = 1; i < argc; i++) {
        char *filename = argv[i];
        int file_fd = open(filename, O_RDONLY);
        if (file_fd == -1) {
            err_exit("open");
            exit(1);
        }
        printf("opened %s file with fd %d\n", filename, file_fd);
        num_open_fds++;
        ev.data.fd = file_fd;
        ev.events = EPOLLIN;
        if (epoll_ctl(epfd, EPOLL_CTL_ADD, file_fd, &ev) == -1)
            err_exit("epoll_ctl");
    }

    char buf[MAXBUF];

    struct epoll_event evlist[MAXEVENTS];
    int ready;
    while (num_open_fds > 0) {
        ready = epoll_wait(epfd, evlist, MAXEVENTS, -1);
        if (ready == -1)
            err_exit("epoll_wait");

        printf("Ready: %d\n", ready);

        for (int j = 0; j < ready; j++) {
            printf(" fd=%d; events: %s%s%s\n", evlist[j].data.fd,
                   (evlist[j].events & EPOLLIN) ? "EPOLLIN " : "",
                   (evlist[j].events & EPOLLHUP) ? "EPOLLHUP " : "",
                   (evlist[j].events & EPOLLERR) ? "EPOLLERR " : "");

            if (evlist[j].events & EPOLLIN) {
                int n = read(evlist[j].data.fd, buf, MAXBUF);
                if (n == -1)
                    err_exit("read");
                printf("    read %d bytes: %.*s\n", n, n, buf);
            } else if (evlist[j].events & (EPOLLHUP | EPOLLERR)) {
                printf("    closing fd %d\n", evlist[j].data.fd);
                if (close(evlist[j].data.fd) == -1)
                    err_exit("close");
                num_open_fds--;
            }
        }
    }

    printf("All file descriptors closed; bye\n");
    exit(EXIT_SUCCESS);
}