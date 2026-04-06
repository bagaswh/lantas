#include <fcntl.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static volatile int watched_fd = -1;

/*
 * SIGHUP handler: reads one byte from watched_fd and prints it.
 * read() and write() are async-signal-safe per POSIX.
 */
static void
sigusr1_handler(int signo)
{
    (void) signo;

    if (watched_fd < 0)
        return;

    unsigned char buf;
    ssize_t n = read(watched_fd, &buf, 1);

    if (n == 1) {
        char msg[64];
        int len = snprintf(msg, sizeof(msg), "byte 0x%02X  '%c'\n", buf,
                           (buf >= 0x20 && buf < 0x7F) ? buf : '.');
        write(STDOUT_FILENO, msg, len);
    } else if (n == 0) {
        const char msg[] = "[EOF]\n";
        write(STDOUT_FILENO, msg, sizeof(msg) - 1);
    } else {
        const char msg[] = "read error\n";
        write(STDERR_FILENO, msg, sizeof(msg) - 1);
    }
}

/*
 * SIGUSR2 handler: fork() a child. Both parent and child share the same
 * open file description (and thus the same offset) since fork() duplicates
 * the fd table. Each subsequent SIGUSR1 in either process advances the
 * shared offset by one byte.
 */
static void
sigusr2_handler(int signo)
{
    (void) signo;

    pid_t pid = fork();

    if (pid < 0) {
        const char msg[] = "fork failed\n";
        write(STDERR_FILENO, msg, sizeof(msg) - 1);
    } else if (pid == 0) {
        /* Child: print its own PID and continue in the signal loop */
        char msg[64];
        int len =
            snprintf(msg, sizeof(msg), "[child]  PID %d forked\n", getpid());
        write(STDOUT_FILENO, msg, len);
    } else {
        /* Parent: print the child's PID */
        char msg[64];
        int len =
            snprintf(msg, sizeof(msg), "[parent] forked child PID %d\n", pid);
        write(STDOUT_FILENO, msg, len);
    }
}

int
main(int argc, char *argv[])
{
    if (argc < 2) {
        fprintf(stderr, "Usage: %s <file>\n", argv[0]);
        fprintf(stderr, "  Send SIGHUP to read and print the next byte.\n");
        return EXIT_FAILURE;
    }

    int fd = open(argv[1], O_RDONLY);
    if (fd < 0) {
        perror("open");
        return EXIT_FAILURE;
    }

    watched_fd = fd;

    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = 0;

    sa.sa_handler = sigusr1_handler;
    if (sigaction(SIGUSR1, &sa, NULL) < 0) {
        perror("sigaction SIGUSR1");
        close(fd);
        return EXIT_FAILURE;
    }

    sa.sa_handler = sigusr2_handler;
    if (sigaction(SIGUSR2, &sa, NULL) < 0) {
        perror("sigaction SIGUSR2");
        close(fd);
        return EXIT_FAILURE;
    }

    printf("PID %d ready.\n", getpid());
    printf("  SIGUSR1 — read and print next byte:  kill -USR1 %d\n",
           getpid());
    printf("  SIGUSR2 — fork a child:               kill -USR2 %d\n",
           getpid());

    /* Just wait for signals forever */
    while (1)
        pause();

    close(fd);
    return EXIT_SUCCESS;
}