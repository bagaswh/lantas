#define _POSIX_C_SOURCE 200809L

#include <arpa/inet.h>
#include <netdb.h>
#include <netinet/in.h>
#include <poll.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

#include "config.h"

#include "third_party/bg/src/container/bg_slice.h"

void *
get_in_addr(struct sockaddr *sa)
{
    if (sa->sa_family == AF_INET)
        return &(((struct sockaddr_in *) sa)->sin_addr);
    return &(((struct sockaddr_in6 *) sa)->sin6_addr);
}

int
get_listener_socket(struct socket_config sockcfg)
{
    int listener;
    int yes = 1;
    int rv;

    struct addrinfo hints, *ai, *p;
    memset(&hints, 0, sizeof hints);
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_flags = AI_PASSIVE;
    char port[7];
    sprintf(port, "%d", sockcfg.port);
    if ((rv = getaddrinfo(NULL, port, &hints, &ai)) != 0) {
        fprintf(stderr, "pollserver: %s\n", gai_strerror(rv));
        exit(1);
    }

    for (p = ai; p != NULL; p = p->ai_next) {
        listener = socket(p->ai_family, p->ai_socktype, p->ai_protocol);
        if (listener < 0)
            continue;

        setsockopt(listener, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(int));

        if (bind(listener, p->ai_addr, p->ai_addrlen) < 0) {
            close(listener);
            continue;
        }

        break;
    }

    freeaddrinfo(ai);

    if (p == NULL)
        return -1;

    if (listen(listener, 10) == -1)
        return -1;

    return listener;
}

BGSlice *
new_pfds()
{
    return BGSlice_new(struct pollfd, 1, 10, NULL);
}

int
add_to_pfds(BGSlice *pfds, int newfd)
{
    struct pollfd pfd = { .fd = newfd, .events = POLLIN };
    if (BGSlice_append(pfds, ((void *) &pfd)) == NULL)
        return -1;
    return 0;
}

void
del_from_pfds(BGSlice *pfds, int i)
{
    // Move the one from the end to this index
    BGSlice_set(pfds, i, BGSlice_get_last(pfds));
    BGSlice_set_len(pfds, BGSlice_get_len(pfds) - 1);
}

int
main(int argc, char **argv)
{
#if (DEBUG == 1)
    setbuf(stdout, NULL);
#endif

    int listener;

    int newfd;
    struct sockaddr_storage remoteaddr;
    socklen_t addrlen;

    char buf[4096];

    char remoteIP[INET6_ADDRSTRLEN];

    int fd_count = 0;
    int fd_size = 5;
    BGSlice *pfds = new_pfds();

    struct socket_config sockcfg = { .reuseaddr = true,
                                     .port = 9118,
                                     .addr = "0.0.0.0" };

    listener = get_listener_socket(sockcfg);

    if (listener == -1) {
        fprintf(stderr, "error getting listening socket\n");
        exit(1);
    }

    printf("Listening at %s:%d", sockcfg.addr, sockcfg.port);

    struct pollfd first_fd = { .fd = listener, .events = POLLIN };
    add_to_pfds(pfds, listener);

    for (;;) {
        int fd_count = BGSlice_get_len(pfds);
        struct pollfd *pfds_ptr = BGSlice_get_data_ptr(pfds);
        int poll_count = poll(pfds_ptr, fd_count, -1);

        if (poll_count == -1) {
            perror("poll");
            exit(1);
        }

        for (int i = 0; i < fd_count; i++) {
            struct pollfd pfd = *((struct pollfd *) BGSlice_get(pfds, i));
            if (pfd.revents & (POLLIN | POLLHUP)) {
                if (pfd.fd == listener) {
                    addrlen = sizeof(remoteaddr);
                    newfd = accept(listener, (struct sockaddr *) &remoteaddr,
                                   &addrlen);
                    if (newfd == -1) {
                        perror("accept");
                    } else {
                        void *remote_in_addr =
                            get_in_addr((struct sockaddr *) &remoteaddr);
                        add_to_pfds(pfds, newfd);
                        printf("new connection from %s on socket %d",
                               inet_ntop(remoteaddr.ss_family, remote_in_addr,
                                         remoteIP, INET6_ADDRSTRLEN),
                               newfd);
                    }
                } else {
                    // BGSlice_range
                }
            }
        }
    }

    return 0;
}