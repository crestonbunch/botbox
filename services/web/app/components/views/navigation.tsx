import * as React from "react";
import * as ReactDOM from "react-dom";
import { observer } from "mobx-react";
import { Icon, IconGroup, Input, Button, Menu, Container } from "semantic-ui-react"

import { Store } from "../../store";
import { LoginPopup } from "../popups/login";
import { NotificationsPopup } from "../popups/notifications";
import { NotificationsIcon } from "../icons/notifications";

export interface NavigationProps {
  store: Store,
}

@observer
export class Navigation extends React.Component<NavigationProps, {}> {

  render() {
    const store = this.props.store

    const notificationItem = (store.notifications.length > 0) ?
      (<NotificationsPopup store={store}
        trigger={
          <Menu.Item name="notifications">
            <NotificationsIcon store={store} />
          </Menu.Item>
        } />) : (
          <Menu.Item name="notifications">
            <NotificationsIcon store={store} />
          </Menu.Item>
        )

    const actionMenu = (store.loggedIn) ? (
      <Menu.Menu position="right">
        {notificationItem}
        <Menu.Item name="user">
          <Icon inverted name="user" />
          {store.session.name}</Menu.Item>
      </Menu.Menu>
    ) : (
        <Menu.Menu position="right">
          <LoginPopup
            store={store}
            trigger={<Menu.Item name='login'>Login</Menu.Item>} />
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

