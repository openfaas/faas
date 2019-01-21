import React, {Component} from 'react'

import SvgIcon from '@material-ui/core/SvgIcon'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import Input from '@material-ui/core/Input'
import Divider from '@material-ui/core/Divider';


export default class SidebarLeft extends Component {
  constructor (props) {
      super(props)

      this.state={
        searchText:'',
        sortKey:'Invocations',
        sortDir:'1',
        selectedFunctionName:'',
        isFunctionBeingCreated:false
      }
  }

  updateSearchText(event){
    event.stopPropgation()
    event.preventDefault()
    this.setState({searchText:event.target.value})
  }

  showFunction(func) {

  }

  createFunction() {

  }

  filterFunctions(func) {
      return func.name.indexOf(this.state.searchText) >= 0
  }

  sortFunctions(funcA,funcB) {
    if (funcA[this.state.sortKey] == funcB[this.state.sortKey]) return 0

    return funcA[this.state.sortKey] > funcB[this.state.sortKey] ? this.state.sortDir : -this.state.sortDir
  }

  render() {

    return (
        <section>
            {this.renderDeployButton()}

            {this.renderSearch()}

            <List>
              { this.renderList() }
            </List>
        </section>
    )
  }

  renderDeployButton() {
    return (
      <List>
        <ListItem className="primary-item" onClick={this.state.isFunctionBeingCreated || this.createFunction.bind(this)}>
          <SvgIcon style={{marginRight: '16px'}} path="img/icons/ic_shop_two_black_24px.svg"> </SvgIcon>
          <p>Deploy New Function</p>
        </ListItem>
      </List>
    )
  }

  renderSearch() {
    if(this.props.functions.length === 0) return null

    return (
      <Input className="md-block">
        <label style="padding-left: 8px">Search for Function</label>
        <input value={this.state.searchText} onChange={this.updateSearchText.bind(this)}></input>
      </Input>
    )
  }

  renderList() {
    return this.props.functions
      .filter(this.filterFunctions.bind(this))
      .sort(this.sortFunctions.bind(this))
      .map(this.renderFunction.bind(this))
  }

  renderFunction(func) {

    //TODO ng-className="function.name == selectedFunction.name ? 'selected' : false">
    //TODO

    return (
      <ListItem className="md-3-line" onClick={this.showFunction.bind(this, func)} >
        <md-icon ng-switch-when="true" style="color: blue" md-svg-icon="person"></md-icon>
        <md-icon ng-switch-when="false" md-svg-icon="person-outline"></md-icon>
        <p>{func.name}</p>
        <Divider ng-if="!$last"></Divider>
      </ListItem>
    )
  }
}
