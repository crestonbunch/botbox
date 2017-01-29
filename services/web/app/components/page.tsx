import * as React from "react";
import { Segment, Container, Header, Image, Button, Divider } from "semantic-ui-react"

import { Navigation } from "../components/navigation";
import { Footer } from "../components/footer";

export class Page extends React.Component<{}, {}> {
  render() {
    return (
    <div>
      <Navigation />
        {this.props.children}
      <Footer />
    </div>
    );
  }
}

