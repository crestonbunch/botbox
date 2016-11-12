import * as React from "react";
import * as ReactDOM from "react-dom";
import { Input, Menu } from "semantic-ui-react"

import { Register } from "./register";
import { Login } from "./login";

export class Navigation extends React.Component<{}, {}> {

  render() {

    return (
      <Menu secondary>
        <Menu.Item name='home' />
        <Menu.Item name='leagues' />
        <Menu.Item name='leaderboard' />

        <Menu.Menu position="right">
          <Login trigger={<Menu.Item name='Login' />}/>
          <Register trigger={<Menu.Item name='Register' />}/>
        </Menu.Menu>
      </Menu>
    );
  }
}

