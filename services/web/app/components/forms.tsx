import * as React from "react";
import { Icon, Label } from "semantic-ui-react";

export interface FieldErrorProps {
  error: string;
}

export class FieldError extends React.Component<FieldErrorProps, {}> {
  render() {
    return this.props.error != "" ? (
      <Label basic pointing="below" color="red">
        <Icon name="warning sign" /> {this.props.error}
      </Label>
    ) : null;
  }
}

