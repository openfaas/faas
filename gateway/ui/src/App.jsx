import React, {Component} from 'react'

import Toolbar from '@material-ui/core/Toolbar'
import Button from '@material-ui/core/Button'
import SvgIcon from '@material-ui/core/SvgIcon'

import SidebarLeft from './components/SidebarLeft.jsx'


export default class OpenFaasUI extends Component {
  constructor(props) {
    super(props)

    this.state = {
      showLeftMenu:true
    }
  }

  toggleLeftMenu() {
    this.setState({
      showLeftMenu:!this.state.showLeftMenu
    })
  }



  render() {
    return (
      <div>
        <div>
          <Toolbar className="md-theme-indigo">
            <div>
              <Button onClick={this.toggleLeftMenu.bind(this)}>
                <SvgIcon 
                  path="img/icons/ic_menu_white.svg"
                > </SvgIcon>
              </Button>
            </div>
          </Toolbar>
          <section>
            <SidebarLeft show={this.state.showLeftMenu} hideMenu={this.toggleLeftMenu.bind(this)}/>
          </section>
        </div>

      </div>
    )
  }
}
