// Copyright 2019 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vfs

import (
	"sync/atomic"

	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/sentry/context"
)

// A Filesystem is a tree of nodes represented by Dentries, which forms part of
// a VirtualFilesystem.
//
// Filesystems are reference-counted. Unless otherwise specified, all
// Filesystem methods require that a reference is held.
//
// Filesystem is analogous to Linux's struct super_block.
type Filesystem struct {
	// refs is the reference count. refs is accessed using atomic memory
	// operations.
	refs int64

	// impl is the FilesystemImpl associated with this Filesystem. impl is
	// immutable. This should be the last field in Dentry.
	impl FilesystemImpl
}

// Init must be called before first use of fs.
func (fs *Filesystem) Init(impl FilesystemImpl) {
	fs.refs = 1
	fs.impl = impl
}

// Impl returns the FilesystemImpl associated with fs.
func (fs *Filesystem) Impl() FilesystemImpl {
	return fs.impl
}

func (fs *Filesystem) incRef() {
	if atomic.AddInt64(&fs.refs, 1) <= 1 {
		panic("Filesystem.incRef() called without holding a reference")
	}
}

func (fs *Filesystem) decRef() {
	if refs := atomic.AddInt64(&fs.refs, -1); refs == 0 {
		fs.impl.Release()
	} else if refs < 0 {
		panic("Filesystem.decRef() called without holding a reference")
	}
}

// FilesystemImpl contains implementation details for a Filesystem.
// Implementations of FilesystemImpl should contain their associated Filesystem
// by value as their first field.
//
// All methods that take a ResolvingPath must resolve the path before
// performing any other checks, including rejection of the operation if not
// supported by the FilesystemImpl. This is because the final FilesystemImpl
// (responsible for actually implementing the operation) isn't known until path
// resolution is complete.
//
// For all methods that take or return linux.Statx, Statx.Uid and Statx.Gid
// should be interpreted as IDs in the root UserNamespace (i.e. as auth.KUID
// and auth.KGID respectively).
//
// FilesystemImpl combines elements of Linux's struct super_operations and
// struct inode_operations, for reasons described in the documentation for
// Dentry.
type FilesystemImpl interface {
	// Release is called when the associated Filesystem reaches zero
	// references.
	Release()

	// Sync "causes all pending modifications to filesystem metadata and cached
	// file data to be written to the underlying [filesystem]", as by syncfs(2).
	Sync(ctx context.Context) error

	// GetDentryAt returns a Dentry representing the file at rp. A reference is
	// taken on the returned Dentry.
	//
	// GetDentryAt does not correspond directly to a Linux syscall; it is used
	// in the implementation of:
	//
	// - Syscalls that need to resolve two paths: rename(), renameat(),
	// renameat2(), link(), linkat().
	//
	// - Syscalls that need to refer to a filesystem position outside the
	// context of a file description: chdir(), fchdir(), chroot(), mount(),
	// umount().
	GetDentryAt(ctx context.Context, rp *ResolvingPath, opts GetDentryOptions) (*Dentry, error)

	// LinkAt creates a hard link at rp representing the same file as vd. It
	// does not take ownership of references on vd.
	//
	// The implementation is responsible for checking that vd.Mount() ==
	// rp.Mount(), and that vd does not represent a directory.
	LinkAt(ctx context.Context, rp *ResolvingPath, vd VirtualDentry) error

	// MkdirAt creates a directory at rp.
	MkdirAt(ctx context.Context, rp *ResolvingPath, opts MkdirOptions) error

	// MknodAt creates a regular file, device special file, or named pipe at
	// rp.
	MknodAt(ctx context.Context, rp *ResolvingPath, opts MknodOptions) error

	// OpenAt returns an FileDescription providing access to the file at rp. A
	// reference is taken on the returned FileDescription.
	OpenAt(ctx context.Context, rp *ResolvingPath, opts OpenOptions) (*FileDescription, error)

	// ReadlinkAt returns the target of the symbolic link at rp.
	ReadlinkAt(ctx context.Context, rp *ResolvingPath) (string, error)

	// RenameAt renames the Dentry represented by vd to rp. It does not take
	// ownership of references on vd.
	//
	// The implementation is responsible for checking that vd.Mount() ==
	// rp.Mount().
	RenameAt(ctx context.Context, rp *ResolvingPath, vd VirtualDentry, opts RenameOptions) error

	// RmdirAt removes the directory at rp.
	RmdirAt(ctx context.Context, rp *ResolvingPath) error

	// SetStatAt updates metadata for the file at the given path.
	SetStatAt(ctx context.Context, rp *ResolvingPath, opts SetStatOptions) error

	// StatAt returns metadata for the file at rp.
	StatAt(ctx context.Context, rp *ResolvingPath, opts StatOptions) (linux.Statx, error)

	// StatFSAt returns metadata for the filesystem containing the file at rp.
	// (This method takes a path because a FilesystemImpl may consist of any
	// number of constituent filesystems.)
	StatFSAt(ctx context.Context, rp *ResolvingPath) (linux.Statfs, error)

	// SymlinkAt creates a symbolic link at rp referring to the given target.
	SymlinkAt(ctx context.Context, rp *ResolvingPath, target string) error

	// UnlinkAt removes the non-directory file at rp.
	UnlinkAt(ctx context.Context, rp *ResolvingPath) error

	// TODO: d_path(); extended attributes; inotify_add_watch(); bind()
}
