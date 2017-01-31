import * as React from "react";
import * as ReactDOM from "react-dom";
import { observer } from "mobx-react";
import { Icon, Input, Button, Menu, Container } from "semantic-ui-react"

import { Store } from "../store";
import { Login } from "./login";

export interface NavigationProps {
  store: Store,
}

@observer
export class Navigation extends React.Component<NavigationProps, {}> {

  render() {
    const store = this.props.store
    const actionMenu = (store.loggedIn) ? (
      <Menu.Menu position="right">
        <Menu.Item name="notifications"><Icon name="alarm" /></Menu.Item>
        <Menu.Item name="user">
          <Icon inverted name="user" /> 
          {store.session.name}</Menu.Item>
      </Menu.Menu>
    ) : (
        <Menu.Menu position="right">
          <Login store={store} trigger={<Menu.Item name='login'>Login</Menu.Item>} />
        </Menu.Menu>
      )

    return (
      <Menu borderless inverted stackable compact fluid>
        <Container>
          <Menu.Item fitted="vertically">
            <img src="assets/botbox-logo-inverted-full.svg" style={{ height: "70%" }} />
          </Menu.Item>
          <Menu.Item name='play'>Play</Menu.Item>
          <Menu.Item name='watch'>Watch</Menu.Item>
          <Menu.Item name='learn'>Learn</Menu.Item>
          {actionMenu}
        </Container>
      </Menu>
    );
  }
}

