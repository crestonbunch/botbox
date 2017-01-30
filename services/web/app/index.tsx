import * as React from "react";
import * as ReactDOM from "react-dom";
import "core-js/shim";

import { Router, Route, Redirect, IndexRedirect, IndexRoute, browserHistory } from 'react-router'

import { Page } from "./components/page";
import { Home } from "./home";
import { Register } from "./register";

import { UserStore } from "./stores/domain/user";
import { RegisterStore } from "./stores/ui/register";

let registerStore = new RegisterStore();
let userStore = new UserStore();

class PageWrapper extends React.Component<{}, {}> {
  render() {
    return <Page userStore={userStore}>{this.props.children}</Page>
  }
}

class RegisterWrapper extends React.Component<{}, {}> {
  render() {
    return <Register registerStore={registerStore} />
  }
}

ReactDOM.render(
  (
    <Router history={browserHistory}>
      <Route path="/" component={PageWrapper}>
        <IndexRoute component={Home} />
        <Route path="register" component={RegisterWrapper} />
      </Route>
    </Router>
  ),
  document.getElementById("app")
);