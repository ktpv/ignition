# ignition web app

The ignition UI is a single page React application.

## Dev Setup

`cd web && yarn install`

### Code formatting

Format the entire project with `yarn run fmt`.

To enable formatting on save in Atom, install the `prettier-atom` package.

Make sure to enable the _ESLint Integration_ checkbox, and disable the
`linter-eslint` package's _Fix on save_ option (to prevent fixing your code twice).

**Note:** we prefer [standard JS style](https://standardjs.com/), which
disagrees with Prettier in one or two ways. In order to make both happy,
we're leveraging `prettier-eslint`, which runs prettier and then feeds its
output into `eslint --fix`. The `prettier-atom` package also takes this approach.

In order to support this flow, we must use `eslint` with a standard plugin rather
than running the `standard` linter directly.
