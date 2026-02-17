#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/select.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/types.h>
#include <unistd.h>

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

    char *filename = argv[1];
    int fd = open(filename, O_RDONLY | O_NONBLOCK, 0644);
    if (fd == -1) {
        fprintf(stderr, "Failed to open file %s: %s\n", filename,
                strerror(errno));
        exit(1);
    }

    fd_set readfds, writefds;
    int ready, nfds, num_read, j;
    struct timeval timeout = { .tv_sec = 0, .tv_usec = 0 };
    char buf[10];

    // Build file descriptor sets
    nfds = 0;
    FD_ZERO(&readfds);
    FD_ZERO(&writefds);
    FD_SET(fd, &readfds);
    FD_SET(fd, &writefds);
    nfds = fd + 1;

    FD_SET(STDIN_FILENO, &writefds);

    for (;;) {
        if ((ready = select(nfds, &readfds, &writefds, NULL, NULL)) == -1)
            err_exit("select");

        for (fd = 0; fd < nfds; fd++) {
            char ready = 0;
            if (FD_ISSET(fd, &readfds))
                ready = 'r';
            if (FD_ISSET(fd, &writefds))
                ready = 'w';
            if (ready > 0)
                printf("%d ready for %c\n", fd, ready);
        }
    }

    exit(EXIT_SUCCESS);
}