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

