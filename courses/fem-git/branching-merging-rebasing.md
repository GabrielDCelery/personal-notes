# Branching, merging, rebasing

We create a new repo with a single file called `A.md` and a branch called `foo` using

```sh
git branch foo
```

Branches are stored in the `refs/heads` folder and are just files pointing to a `sha`. After creating the branch the `sha` that it is pointing to is going to be the same of the `sha` we branched off of.

```sh
cat ./.git/refs/heads/trunk
cc1d300b34c7ca6ea3763da17d3482badc46413sh

cat ./.git/refs/heads/foo
cc1d300b34c7ca6ea3763da17d3482badc46413sh
```

### Merge

Merge is there to attempt to combine two histories together that diverged from each other at some point. This is done by using the `best common ancestor` (or also called as `merge base`) which is just a fancy term for the earliest `sha` that are the same on both branches.

Once it find a common ancestor then it applies all the changes on top of each other and creates an extra commit (the `merge commit`) that gets applied on top of the whole thing.

You execute the command from the `current branch` you are on and specify the branch `you want to merge into current`.

```sh
git merge <targetbranch> # This will merge in things from target branch to the current branch you are on
```

We can see how the merge strategy was applied using the `--grap` and `--parent` flag.

```sh
git log --graph --oneline --parent

*   286fc98 5ca3742 2440841 merge branch 'foo' into trunk-merge-foo
|\  
| * 2440841 86bc7f8 c
| * 86bc7f8 8ec7767 b
* | 5ca3742 c7e23c2 e
* | c7e23c2 8ec7767 d
|/  
* 8ec7767 a
```

If the `best common ancestor` is the tip of the branch you are merging into then no new commit message will be created, all that will happen is that the divergent branch will be put on top of the branch you are merging into. This is what is called `fast forwarding`. This is because you just have a linear history with no divergence.

### Rebase

The basic steps of rebase

1. go to the branch that you want to rebase -> `current branch`
2. run `git rebase <target branch>`
3. this will check out the latest commit on `target branch`
4. and then will play one commit at a time from `current branch`
5. once finished will update `curent branch` to the `commit sha`

```sh
git log --oneline --graph --decorate --parents

* d8e8d0c 5a0f61a (head -> foo-rebase-trunk) c
* 5a0f61a 7f9035b b
* 7f9035b c290570 (trunk) y
* c290570 5ca3742 x
* 5ca3742 c7e23c2 e
* c7e23c2 8ec7767 d
* 8ec7767 a
```

Unlike with merge now we have a nice clean linear history and we can `fast forward` into the branch that we used for our rebase.

It is important that rebase does not rewrite the history of the branch you are rebasing on the top of, but does it to the commits that get rebased.

```sh
# This history was not rewritten

* 7f9035b (HEAD -> trunk) Y
* c290570 X
* 5ca3742 E
* c7e23c2 D
* 8ec7767 A

# This history was rewritten
* 2440841 (HEAD -> foo) C
* 86bc7f8 B
* 8ec7767 A
```

### HEAD and Reflog

in the `.git` folder there is a file named `HEAD` that stores a pointer to a ref, which stores the `sha` of a commit.

```sh
cat HEAD
ref: refs/heads/trunk # This is the file HEAD is pointing to

cat refs/heads/foo
244084138563fe6439b41f88b971f283ff478611 # This is the sha the reference is pointing to
```

`Reflog` is just the history of where `HEAD` had been.

```sh
git reflog

7f9035b HEAD@{0}: checkout: moving from foo to trunk
2440841 HEAD@{1}: checkout: moving from trunk to foo
7f9035b HEAD@{2}: checkout: moving from foo to trunk
2440841 HEAD@{3}: checkout: moving from trunk to foo
7f9035b HEAD@{4}: checkout: moving from foo-rebase-trunk to trunk
d8e8d0c HEAD@{5}: rebase (finish): returning to refs/heads/foo-rebase-trunk
d8e8d0c HEAD@{6}: rebase (pick): C
5a0f61a HEAD@{7}: rebase (pick): B
7f9035b HEAD@{8}: rebase (start): checkout trunk
2440841 HEAD@{9}: checkout: moving from foo to foo-rebase-trunk
2440841 HEAD@{10}: checkout: moving from trunk to foo
7f9035b HEAD@{11}: commit: Y
c290570 HEAD@{12}: commit: X
5ca3742 HEAD@{13}: checkout: moving from foo-merge-trunk to trunk
bf06174 HEAD@{14}: checkout: moving from trunk-merge-foo to foo-merge-trunk
286fc98 HEAD@{15}: checkout: moving from trunk-merge-foo to trunk-merge-foo
286fc98 HEAD@{16}: checkout: moving from foo-merge-trunk to trunk-merge-foo
bf06174 HEAD@{17}: checkout: moving from trunk-merge-foo to foo-merge-trunk
286fc98 HEAD@{18}: checkout: moving from foo-merge-trunk to trunk-merge-foo
bf06174 HEAD@{19}: merge trunk: Merge made by the 'ort' strategy.
2440841 HEAD@{20}: checkout: moving from foo to foo-merge-trunk
2440841 HEAD@{21}: checkout: moving from trunk to foo
5ca3742 HEAD@{22}: checkout: moving from trunk-merge-foo to trunk
286fc98 HEAD@{23}: merge foo: Merge made by the 'ort' strategy.
5ca3742 HEAD@{24}: checkout: moving from trunk to trunk-merge-foo
5ca3742 HEAD@{25}: commit: E
c7e23c2 HEAD@{26}: commit: D
8ec7767 HEAD@{27}: checkout: moving from foo to trunk
2440841 HEAD@{28}: checkout: moving from trunk to foo
8ec7767 HEAD@{29}: checkout: moving from foo to trunk
2440841 HEAD@{30}: commit: C
86bc7f8 HEAD@{31}: commit: B
8ec7767 HEAD@{32}: checkout: moving from trunk to foo
8ec7767 HEAD@{33}: commit (initial): A
```

### Merging individual sha

One of the great things git allows you to do is not just to merge branches, bur individual commits as well. For that you need the specific `sha` and rather then specifying the name of the branch you use that.

```sh
git merge <sha>
```

This can also be used to restroe branches that you accidentally deleted. Assume you ran a `gir branch -D baz` and you just deleted a branch of yours. You could do the following:

```sh
git reflog -5 # Use this to grab the latest commits


7f9035b HEAD@{0}: checkout: moving from baz to trunk
53c8298 HEAD@{1}: commit: baz
7f9035b HEAD@{2}: checkout: moving from trunk to baz
7f9035b HEAD@{3}: checkout: moving from foo to trunk
2440841 HEAD@{4}: checkout: moving from trunk to foo

git merge 53c8298
```

The only issue with that if in the meantime the branch you are merging into has changed too then all the changes since then will be applied from the `common ancestor`. So if we want to avoid that then better to use `cherry pick`.

With the above example we would do:

```sh
git cherry-pick 53c8298
```
