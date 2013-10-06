
/*
	Dex provides a mechanism for storing filesystem images of the kind used by docker, using git as a datastore/deduplication/history/transport system.

	Git is used fairly directly and naturally by dex, so it remains very possible to use normal tools to both inspect and modifying the repository:
	 - Normal git history viewers like `git log`, gitg, gitk, etc will show the graph of which images were used to produce which others, when, and so on.
	 - `git checkout` can give you any raw image.tar you want to review.
	 - Moving a branch reference to point to a different commit lets you tell dex what you want that image name to mean, because that's all the branch is.
	 - `git branch -D` can drop any branch of images, and they disappear, and the disk space is reclaimable when you `git gc`.
*/
package dex