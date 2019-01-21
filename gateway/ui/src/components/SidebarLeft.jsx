import React, {Component} from 'react'

import Toolbar from '@material-ui/core/Toolbar'
import Button from '@material-ui/core/Button'
import SvgIcon from '@material-ui/core/SvgIcon'
import FunctionList from './FunctionList.jsx'


export default class SidebarLeft extends Component {

  render() {
    return (
      <div className="md-sidenav-left">
        <Toolbar className="md-theme-indigo">
          <div className="md-toolbar-tools">
            <a href="https://www.openfaas.com/" target="_blank"><img src="icon.png" alt="OpenFaaS Icon" width="60px" height="60px" className="avatar" /></a>
            <h1 style={{flex: '1 1 auto'}}>&nbsp; OpenFaaS Portal</h1>
              <Button className="md-icon-button" aria-label="Back" onClick={this.props.hideMenu}>
                <SvgIcon 
                  path="img/icons/ic_arrow_back_white.svg"
                > </SvgIcon>
              </Button>
          </div>
        </Toolbar>

        <FunctionList functions={[]} handleDeployFunction={()=>{}}/>

      </div>
    )
  }
}
