import * as React from "react";
import * as ReactDOM from "react-dom";
import "core-js/shim";

import { Router, Route, Link, hashHistory } from 'react-router'

import { Home } from "./pages/home";
import { GithubLogin } from "./pages/github";

ReactDOM.render(
(
    <Router history={hashHistory}>
      <Route path="/" component={Home} />
      <Route path="/github" component={GithubLogin} />
    </Router>
),
    document.getElementById("app")

);
