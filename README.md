## Clip

A set of tools to make managing remote git branches easy and fun

### git clip

Due to the collaborative nature of git; over time one begins to accumulate
quite a few branches local and remote branches. Keeping track of what branch
needs to be merged, has already been merged, and what branches need a pull can
tax the little grey cells (aka the brain).

``git clip`` is designed to help eleviate this pain by providing a clear view
of all your local and remote branches.

![alt tag](https://raw.githubusercontent.com/thrawn01/clip/master/gifs/clip.gif)

### git clip-remote
Over time forked repo's can collect a large number of branches left on 'origin'.
``clip-remote`` makes cleaning up these branches simple. It will only ever ask
you to delete branches that are on the remote, but not local and will never ask
to delete tracked branches even if the local branch name differs from the remote
 branch name.

![alt tag](https://raw.githubusercontent.com/thrawn01/clip/master/gifs/clip-remote.gif)


