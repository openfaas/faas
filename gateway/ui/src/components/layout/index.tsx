import React, { SFC, ReactNode } from "react";
import { MuiThemeProvider, createMuiTheme } from "@material-ui/core/styles";
import CssBaseline from "@material-ui/core/CssBaseline";
import { makeStyles } from "@material-ui/styles";

import SideBar, { sidebarWidth } from "./sidebar";

const theme = createMuiTheme({
  palette: {
    type: "light"
  }
});

const styles = makeStyles({
  Inner: {
    width: `calc(100% - ${sidebarWidth}px)`,
    marginLeft: sidebarWidth
  }
});

interface Props {
  children: ReactNode;
}

const Layout: SFC<Props> = ({ children }) => {
  const classes = styles();
  return (
    <MuiThemeProvider theme={theme}>
      <CssBaseline />
      <SideBar />
      <div className={classes.Inner}>{children}</div>
    </MuiThemeProvider>
  );
};

export default Layout;
