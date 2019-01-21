import React from 'react'
import ReactDOM from 'react-dom'
import App from './App.jsx'

import indigo from '@material-ui/core/colors/indigo'
import { MuiThemeProvider,createMuiTheme } from '@material-ui/core/styles'
import CssBaseline from '@material-ui/core/CssBaseline';

const theme = createMuiTheme({
  palette: {
    primary: indigo,
  },
})

function ThemedApp() {
    return (
      <MuiThemeProvider theme={theme}>
        <CssBaseline />
        <App />
      </MuiThemeProvider>
    );
  }


ReactDOM.render(<ThemedApp />, document.getElementById('app'))
