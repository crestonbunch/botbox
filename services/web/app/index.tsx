import * as React from "react";
import * as ReactDOM from "react-dom";
import "core-js/shim";

import { Router, Route, Redirect, IndexRedirect, IndexRoute, browserHistory } from 'react-router'

import { Page } from "./components/views/page";
import { HomePage } from "./components/pages/home";
import { RegistrationPage } from "./components/pages/registration";

import { Store } from "./store";

let store = new Store();

ReactDOM.render(
  (
    <Router history={browserHistory}>
      <Route path="/" component={(props) => { return <Page store={store} {...props} />}}>
        <IndexRoute component={HomePage} />
        <Route path="register" component={() => {return <RegistrationPage store={store} />}} />
      </Route>
    </Router>
  ),
  document.getElementById("app")
);