import * as React from "react";
import * as ReactDOM from "react-dom";
import "core-js/shim";

import { Router, Route, Link, hashHistory } from 'react-router'

import { Home } from "./components/home";
import { Register } from "./components/register";

ReactDOM.render(
(
    <Router history={hashHistory}>
      <Route path="/" component={Home} />
      <Route path="/register" component={Register}/>
    </Router>
),
    document.getElementById("app")

);
