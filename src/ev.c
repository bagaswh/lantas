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
    int file_fd = open(filename, O_RDONLY | O_NONBLOCK, 0644);
    if (file_fd == -1) {
        fprintf(stderr, "Failed to open file %s: %s\n", filename,
                strerror(errno));
        exit(1);
    }

    fd_set readfds, writefds;
    int ready, nfds, num_read, j;
    struct timeval timeout = { .tv_sec = 0, .tv_usec = 0 };
    char buf[10];

    for (;;) {
        FD_ZERO(&readfds);
        FD_ZERO(&writefds);
        FD_SET(file_fd, &readfds);
        FD_SET(STDIN_FILENO, &readfds);
        nfds = file_fd + 1;
        if ((ready = select(nfds, &readfds, &writefds, NULL, NULL)) == -1)
            err_exit("select");

        for (int fd = 0; fd <= nfds; fd++) {
            char fd_ready = 0;
            if (FD_ISSET(fd, &readfds))
                fd_ready = 'r';
            if (FD_ISSET(fd, &writefds))
                fd_ready = 'w';
            if (ready > 0)
                printf("%d ready for %c\n", fd, fd_ready);

            if (fd_ready == 'r') {
                num_read = read(file_fd, buf, 10);
                if (num_read == -1)
                    err_exit("read");
                if (num_read == 0)
                    break;
                printf("Read %d bytes: %s\n", num_read, buf);
            } else if (fd_ready == 'w') {
                printf("Write to stdin\n");
                j = write(fd, "Hello World\n", 13);
                if (j == -1)
                    err_exit("write");
            }
        }
    }

    exit(EXIT_SUCCESS);
}