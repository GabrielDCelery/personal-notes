# Git Basics

### Review or explore

To view commands use `git log`.

```sh
$ git log

commit b6849226c7e534380c21ec6aba03296c1d4b11df
Author: GaborZeller <gabriel.d.celery@gmail.com>
Date:   Fri Aug 23 20:22:22 2024 +0100

    Initial commit
```

Use the `--graph` to show the node tree of the commits and `--decorate` to print out ref names. The latter is important if you want to write all the information of the commit to file and don't want to lose the tags.

```sh
git log --graph --decorate

* commit b6849226c7e534380c21ec6aba03296c1d4b11df (HEAD -> master)
  Author: GaborZeller <gabriel.d.celery@gmail.com>
  Date:   Fri Aug 23 20:22:22 2024 +0100
  
      Initial commit
```


### Finding the sha in the .git folder

If you have the following commit.

```sh
git log --graph --decorate

* commit b6849226c7e534380c21ec6aba03296c1d4b11df (HEAD -> master)
  Author: GaborZeller <gabriel.d.celery@gmail.com>
  Date:   Fri Aug 23 20:22:22 2024 +0100
  
      Initial commit
```

then you can search for the sha in the `.git` folder.

```sh
grep -r b6849226c7e534380c21ec6aba03296c1d4b11df . # This command is initiated from within the .git folder

./refs/heads/master:b6849226c7e534380c21ec6aba03296c1d4b11df
./logs/HEAD:0000000000000000000000000000000000000000 b6849226c7e534380c21ec6aba03296c1d4b11df GaborZeller <gabriel.d.celery@gmail.com> 1724440942 +0100	commit (initial): Initial commit
./logs/refs/heads/master:0000000000000000000000000000000000000000 b6849226c7e534380c21ec6aba03296c1d4b11df GaborZeller <gabriel.d.celery@gmail.com> 1724440942 +0100	commit (initial): Initial commit
```

or

```sh
find . -name "b6*"

./objects/b6
```


### The ./objects folder

The `./objects` folder stores all of the changes. It contains a bunch of subfolders where each folder starts with the first two letters of the sha.

```sh
cat ./objects/b6/849226c7e534380c21ec6aba03296c1d4b11df
```


### How to find the snapshot of what we committed

There is a useful command that can be used to drill down from the `sha` all the way to the project that was committed.

```sh
git cat-file -p b6849226c7e534380c21ec6aba03296c1d4b11df

tree 7edd9b125e2cc50b74f1323a2e1af4e98d84b00b
author GaborZeller <gabriel.d.celery@gmail.com> 1724440942 +0100
committer GaborZeller <gabriel.d.celery@gmail.com> 1724440942 +0100

Initial commit
```

```sh
git cat-file -p 7edd9b125e2cc50b74f1323a2e1af4e98d84b00b

100644 blob 303ff981c488b812b6215f7db7920dedb3b59d9a	first.md
```

```sh
git cat-file -p 303ff981c488b812b6215f7db7920dedb3b59d9a

<contents of the first file that was committed>
```

In  the above example following the breadcrumbs we used the `sha` to find a `tree` which is just a fancy word for a `directory/folder` and then we followed that to find a `blob` which is just a `file` and that led us to the actual contents of the file we committed.
