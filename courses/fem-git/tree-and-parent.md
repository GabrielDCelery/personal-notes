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



###
