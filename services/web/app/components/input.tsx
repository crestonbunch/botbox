import * as React from "react";
import { observer } from "mobx-react";
import { Icon, Label, Form, Button} from "semantic-ui-react";

export interface InputProps {
  type: string,
  label: string,
  error: string,
  placeholder: string,
  onChange: (val: string) => void,
}

export interface InputState {
}

@observer
export class Input extends React.Component<InputProps, InputState> {

  constructor(props: InputProps) {
    super(props);
  }

  render() {
    return (
      <Form.Field error={this.props.error != ""}>
        <label>{this.props.label}</label>
        {
          (this.props.error) ? 
          (
            <Label basic pointing="below" color="red">
              <Icon name="warning sign" /> {this.props.error}
            </Label>
          ) : (<span></span>)
        }
        <input type={this.props.type} 
               placeholder={this.props.placeholder} 
               onChange={(e) => this.props.onChange(e.target.value)} />
      </Form.Field>
    )
  }
}