import * as React from "react";
import * as ReactDOM from "react-dom";
import "core-js/shim";

import { Router, Route, Redirect, IndexRedirect, IndexRoute, browserHistory } from 'react-router'

import { Page } from "./components/page";

import { Home } from "./home";
import { Register } from "./register";

ReactDOM.render(
  (
    <Router history={browserHistory}>
      <Route path="/" component={Page}>
        <IndexRoute component={Home} />
        <Route path="register" component={Register} />
      </Route>
    </Router>
  ),
  document.getElementById("app")
);