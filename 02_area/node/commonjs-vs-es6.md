---
title: "CommonJs vs ES6"
tags: ["node", "commonjs", "es6"]
---

# The problem

Wanted to have some refresher around the difference between `CommonJS` and `NodeJS`.

# The history

## Originally Javascript was working in the browser only

Javascript in the browser was imported using the `script` tag and there was no module system so variables were shared globally. These were called using the `EX` naming convention, e.g. `ES3 (1999-2009)`, `ES5 (2009-2015)`.

```txt
<!-- index.html -->
<script src="utils.js"></script>
<script src="app.js"></script>
```

```js
// utils.js
var myUtils = {
  add: function (a, b) {
    return a + b;
  },
};

// app.js
// Just access the global variable
console.log(myUtils.add(1, 2));
```

## 2009 - Node.js was created

It needed something immediately so we got CommonJS that was using the `require` syntax. It was simple and it just worked.

People tried to make CommonJS work in the browsers using tools like `Browserify` and `Webpack` that parsed commonjs code into bundles that would work in the browser, and it worked-ish but was still far from ideal.

## 2015 - ES modules became the official standard

`ES6` marked a milestone in unifying the browser and backend. Since it was released in `2015` people also refer to it as `ES2015`.

After becoming the standard `browsers` started supporting it around `2017`.

```txt
<script src="old-library.js"></script> <!-- Regular script, still using global scope -->
<script type="module" src="app.js"></script> <!-- ES6 module -->
```

`NodeJS` also started using it with the experimental flags first. Then starting with `nodejs 14` in the `package.json` you had to set the `"type": "module"` filed to tell nodejs how to interpret the `js` files (or you had to use `.mjs` or `.cjs` extensions) and starting with `nodejs 20` you can use either or without the `package.json` and nodejs will be smart enough to use the proper solution.

The naming convnetion sill follows thie followng standard:

```txt
ES6 = ES2015 (they're the same thing)
ES7 = ES2016
ES8 = ES2017
And so on...
```

> [!NOTE]
> Nobody really says ES7, ES11 etc... anymore, people will say "I am writing ES6" to refer to a modern ESwhatever syntax

# NodeJS version vs ES version

Since ESX is the specificaion for the language's capabilities and Nodejs is a runtime for executing that code NodeJS is constantly evolving but it should be always cheked which version of NodeJs can execute which version of ES specification.

Quick Reference Table:

| ES Version   | Minimum Node.js | Released |
| ------------ | --------------- | -------- |
| ES2015 (ES6) | Node.js 6       | 2016     |
| ES2016       | Node.js 7       | 2017     |
| ...          |                 |          |
| ES2023       | Node.js 20      | 2023     |
| ES2024       | Node.js 22      | 2024     |

# Typescript

When transpiling typescript code in the `tsconfig.json` file we have to set the `"target"` (what TO are we compiling) and the `"module"` (what FROM are we compiling) fields. These have a bunch of namings.

## TARGET

This controls what JavaScript features are used in the compiled output:

| Target        | Year      | What you get                                       |
| ------------- | --------- | -------------------------------------------------- |
| es3           | 1999      | Ancient, don't use                                 |
| es5           | 2009      | No arrow functions, classes, let/const - super old |
| es6/es2015    | 2015      | Arrow functions, classes, let/const, promises      |
| es2016        | 2016      | + \*\* operator, Array.includes                    |
| es2017        | 2017      | + async/await                                      |
| es2018        | 2018      | + rest/spread for objects                          |
| es2019        | 2019      | + Array.flat, Object.fromEntries                   |
| es2020        | 2020      | + optional chaining ?., nullish coalescing ??      |
| es2021-es2024 | 2021-2024 | Newer features each year                           |
| esnext        | Latest    | Bleeding edge, all latest features                 |

## MODULE

- Module systems you'll actually use

| Module   | What it is                              | When to use                                  |
| -------- | --------------------------------------- | -------------------------------------------- |
| commonjs | Node.js legacy (require/module.exports) | Old Node.js projects                         |
| esnext   | ES modules, latest syntax               | Modern projects (browser or Node.js)         |
| nodenext | Follows Node.js rules exactly           | Node.js projects that need dual ESM/CommonJS |

- Module systems you probably won't use

| Module               | What it is                             | When to use                                           |
| -------------------- | -------------------------------------- | ----------------------------------------------------- |
| none                 | No module system                       | Tiny scripts with no imports                          |
| amd                  | Async Module Definition (RequireJS)    | Legacy browser code, basically dead                   |
| umd                  | Universal Module Definition            | Libraries supporting both AMD and CommonJS - outdated |
| system               | SystemJS loader                        | Very rare, old bundler                                |
| es2015/es2020/es2022 | ES modules with specific syntax limits | Rarely needed, just use esnext                        |
| node16/node18/node20 | Locked to specific Node.js versions    | Use nodenext instead                                  |
| preserve             | Don't transform imports at all         | Bundlers like Bun that handle it themselves           |

## Additional issues

There is also the added issue of what is it that typescript CAN compile and what is it that it CAN NOT.

1. Syntax features - TypeScript CAN transpile these ✅

Examples:

- Arrow functions (() => {})
- Classes
- Async/await

2. Runtime APIs - TypeScript CANNOT transpile these ❌

Examples:

- Array.includes() (ES2016)
- Object.entries() (ES2017)
- Promise (ES2015)
- Array.flat() (ES2019)

> [!WARN]
> so you will either need a `polyfil` or configure the `lib` property in `tsconfig.json`

## Instead of guessing, use official presets

```sh
  npm install --save-dev @tsconfig/node18

```

```json
// this is for the tsconfig.json
{
  "extends": "@tsconfig/node18/tsconfig.json",
  "compilerOptions": {
    "module": "nodenext"
  }
}
```
