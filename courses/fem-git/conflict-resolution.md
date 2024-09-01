# Conflict resolution

## Options of dealing with conflict resolution

The problem we all face is that after making a change we might need to integrate someone else's change into our codebase. There are usually several options to do that:

- stash your changes
- commit your changes then rebase
- use worktrees

### Stash

Stash takes the changes of your `index` (the staging area) and your `work tree` and puts them into a special place for safekeeping until you need it. The stash is basically a `stack`. The intention here is to revert the repo `to match the HEAD commit` without losing the changes that you made so far.

```sh
git stash # The command to stash your changes
git stash -m "some message" # Your friendly message to do the stashing with

git stash list # List your stashed items
git stash show [--index <index>] # This will show the changes at a current index

git stash pop # pop the latest changes on top of the current working tree
git stash pop --index <index> # pop a specific index on top of the current working tree
```




