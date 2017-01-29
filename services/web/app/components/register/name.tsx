import * as React from "react";
import "whatwg-fetch"
import { Label, Message, Input, Icon, Form, Button } from "semantic-ui-react";

const NAME_LENGTH = 20;

export interface NameProps {
}

export interface NameState {
  error?: string
  value?: string
}

export class Name extends React.Component<NameProps, NameState> {

  constructor(props: NameProps) {
    super(props);
    this.state = {
      value: "",
    };
  }

  getValue(): string {
    return this.state.value;
  }

  updateName(e: Event) {
    let value = (e.target as any).value;

    this.setState({
      value: value,
      error: undefined
    }, this.validate);
  }

  validate(): Promise<boolean> {
    if (this.state.value == "") {
      this.setState({
        error: "Please provide a name."
      });
      return new Promise<boolean>(function(r) {r(false);});
    } else if (this.state.value.length > NAME_LENGTH) {
      this.setState({
        error: "Name must be no more than 20 characters."
      });
      return new Promise<boolean>(function(r) {r(false);});
    }

    return new Promise<boolean>(function(r) {r(true);});
  }

  render() {
    return (
    <Form.Field required error={this.state.error != undefined}>
      <label>Display name</label>
      {this.state.error != undefined ?
        <Label basic pointing="below" color="red"><Icon name="warning sign" />
          {this.state.error}
        </Label>
      : null}
      <Input icon='users' iconPosition='left'
        onChange={this.updateName.bind(this)}
        name="name" placeholder="Enter a display name" />
    </Form.Field>
    )
  }

}
