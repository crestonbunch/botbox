import * as React from "react";
import * as ReactDOM from "react-dom";
import "core-js/shim";

import { Router, Route, Redirect, IndexRedirect, IndexRoute, browserHistory } from 'react-router'

import { Page } from "./components/page";
import { Home } from "./home";
import { Register } from "./register";

import { Store } from "./store";

let store = new Store();

class PageWrapper extends React.Component<{}, {}> {
  render() {
    return <Page store={store}>{this.props.children}</Page>
  }
}

ReactDOM.render(
  (
    <Router history={browserHistory}>
      <Route path="/" component={PageWrapper}>
        <IndexRoute component={Home} />
        <Route path="register" component={() => {return <Register store={store} />}} />
      </Route>
    </Router>
  ),
  document.getElementById("app")
);