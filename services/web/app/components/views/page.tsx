import * as React from "react";
import { Icon, Message, Container, Header, Image, Button, Divider } from "semantic-ui-react"

import { Navigation } from "./navigation";
import { Footer } from "./footer";
import { Store } from "../../store"

export interface PageProps {
  store: Store;
}

export class Page extends React.Component<PageProps, {}> {

  constructor(props: PageProps) {
    super(props);
  }
  
  render() {
    return (
      <div>
        <Navigation store={this.props.store} />
        {this.props.children}
        <Footer />
      </div>
    );
  }
}

