load("@io_bazel_rules_go//go:def.bzl", "go_test")

package(licenses = ["notice"])

load("//tools/go_stateify:defs.bzl", "go_library")

go_library(
    name = "eventfd",
    srcs = ["eventfd.go"],
    importpath = "gvisor.dev/gvisor/pkg/sentry/kernel/eventfd",
    visibility = ["//pkg/sentry:internal"],
    deps = [
        "//pkg/abi/linux",
        "//pkg/fdnotifier",
        "//pkg/sentry/context",
        "//pkg/sentry/fs",
        "//pkg/sentry/fs/anon",
        "//pkg/sentry/fs/fsutil",
        "//pkg/sentry/usermem",
        "//pkg/syserror",
        "//pkg/waiter",
    ],
)

go_test(
    name = "eventfd_test",
    size = "small",
    srcs = ["eventfd_test.go"],
    embed = [":eventfd"],
    deps = [
        "//pkg/sentry/context/contexttest",
        "//pkg/sentry/usermem",
        "//pkg/waiter",
    ],
)
