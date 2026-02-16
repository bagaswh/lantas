#include <unistd.h>

struct socket_config {
    bool reuseaddr;
    int port;
    char *addr;
};