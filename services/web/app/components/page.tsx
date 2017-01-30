import * as React from "react";
import { Icon, Message, Container, Header, Image, Button, Divider } from "semantic-ui-react"

import { UserStore } from "../stores/domain/user"
import { Navigation } from "./navigation";
import { Footer } from "./footer";

export interface PageProps {
  userStore: UserStore;
}

export class Page extends React.Component<PageProps, {}> {

  constructor(props: PageProps) {
    super(props);
  }
  
  render() {
    return (
      <div>
        <Navigation userStore={this.props.userStore} />
        {this.props.children}
        <Footer />
      </div>
    );
  }
}

