import * as React from "react";
import 'grecaptcha'
import { Label, Message, Input, Icon, Form, Button } from "semantic-ui-react";
import { Settings } from "../../settings";

const RECAPTCHA_ID = "recaptcha";

export interface RecaptchaProps {
}

export interface RecaptchaState {
  ready?: boolean,
  readyCheck?: any,
  error?: string,
  value?: string
}

export class Recaptcha extends React.Component<RecaptchaProps, RecaptchaState> {

  constructor(props: RecaptchaProps) {
    super(props);
    this.state = {
      ready: false,
      readyCheck: setInterval(this.checkReady.bind(this), 200),
      value: ""
    }
  }

  private renderRecaptcha() {
    if (this.state.ready) {
      this.forceUpdate();
      grecaptcha.render(
        RECAPTCHA_ID,
        {
          sitekey: Settings.RECAPTCHA_KEY,
          callback: this.updateRecaptcha.bind(this),
        }
      );
    }
  }

  private checkReady() {
    if (grecaptcha != undefined) {
      this.setState({
        ready: true,
      });
      this.renderRecaptcha();
      clearInterval(this.state.readyCheck);
    }
  }

  getValue(): string {
    return this.state.value;
  }

  updateRecaptcha(value: string) {
    this.setState({
      value: value,
      error: undefined
    });
  }

  validate(): Promise<boolean> {
    if (this.state.value == "") {
      this.setState({
        error: "Please answer the captcha."
      });
      return new Promise<boolean>(function(r) {r(false);});
    }

    return new Promise<boolean>(function(r) {r(true);});
  }


  reset() {
    if (grecaptcha != undefined) {
      grecaptcha.reset();
    }
  }

  render() {
    if (!this.state.ready) {
      return <Icon size="large" name="notched circle" loading />
    }
    return (
      <Form.Field required error={this.state.error != undefined}>
        <label>Singularity check</label>
        {this.state.error != undefined ? <Label basic pointing="below" color="red">{this.state.error}</Label> : null}
        <div id={RECAPTCHA_ID} data-sitekey={Settings.RECAPTCHA_KEY}></div>
      </Form.Field>
    );
  }
}

