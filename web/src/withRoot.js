import React from 'react'
import { MuiThemeProvider, createMuiTheme } from 'material-ui/styles'
import teal from 'material-ui/colors/teal'
import blue from 'material-ui/colors/blue'
import CssBaseline from 'material-ui/CssBaseline'

const theme = createMuiTheme({
  palette: {
    primary: teal,
    secondary: blue,
    type: 'light'
  }
})

function withRoot (Component) {
  function WithRoot (props) {
    return (
      <MuiThemeProvider theme={theme}>
        <CssBaseline />
        <Component {...props} />
      </MuiThemeProvider>
    )
  }

  return WithRoot
}

export default withRoot
