import * as React from "react";
import * as ReactDOM from "react-dom";
import { Input, Button, Menu, Container } from "semantic-ui-react"

import { Login } from "./login";

export class Navigation extends React.Component<{}, {}> {

  render() {

    return (
      <Menu borderless inverted stackable compact fluid>
        <Container>
          <Menu.Item fitted="vertically">
            <img src="assets/botbox-logo-inverted-full.svg" style={{height:"70%"}} />
          </Menu.Item>
          <Menu.Item name='play'>Play</Menu.Item>
          <Menu.Item name='watch'>Watch</Menu.Item>
          <Menu.Item name='learn'>Learn</Menu.Item>

          <Menu.Menu position="right">
            <Login trigger={<Menu.Item name='login'>Login</Menu.Item>}/>
          </Menu.Menu>
        </Container>
      </Menu>
    );
  }
}

