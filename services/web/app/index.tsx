import * as React from "react";
import * as ReactDOM from "react-dom";
import "core-js/shim";

import { Router, Route, Redirect, IndexRedirect, IndexRoute, browserHistory } from 'react-router'

import { Page } from "./components/page";
import { Home } from "./home";

import { Register } from "./register";
import { RegisterStore } from "./stores/ui/register";

let registerStore = new RegisterStore();

class RegisterWrapper extends React.Component<{}, {}> {
  render() {
    return <Register store={registerStore} />
  }
}

ReactDOM.render(
  (
    <Router history={browserHistory}>
      <Route path="/" component={Page}>
        <IndexRoute component={Home} />
        <Route path="register" component={RegisterWrapper} />
      </Route>
    </Router>
  ),
  document.getElementById("app")
);