import * as React from "react";
import * as ReactDOM from "react-dom";
import { observer } from "mobx-react";
import { Icon, Input, Button, Menu, Container } from "semantic-ui-react"

import { UserStore } from "../stores/domain/user";
import { LoginStore } from "../stores/ui/login";
import { Login } from "./login";

export interface NavigationProps {
  userStore: UserStore,
}

@observer
export class Navigation extends React.Component<NavigationProps, {}> {

  loginStore: LoginStore;

  constructor(props: NavigationProps) {
    super(props);
    this.loginStore = new LoginStore(this.props.userStore);
  }

  render() {
    const user = this.props.userStore
    const actionMenu = (user.loggedIn) ? (
      <Menu.Menu position="right">
        <Menu.Item name="notifications"><Icon name="alarm" /></Menu.Item>
        <Menu.Item name="user">
          <Icon inverted name="user" /> 
          {user.session.name}</Menu.Item>
      </Menu.Menu>
    ) : (
        <Menu.Menu position="right">
          <Login trigger={<Menu.Item name='login'>Login</Menu.Item>}
            loginStore={this.loginStore} />
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

