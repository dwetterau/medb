import React, { Component } from 'react';
import PagedrawGeneratedPage from './pagedraw/editfile'
import $ from 'jquery';
import {handlePull, handlePush} from './Api'

class EditFile extends Component {
  constructor(props) {
      super();
      this.state = {
          fileContent: props.file.content,
      }
  }

  handleContentChange(e) {
      this.setState({
          fileContent: e.target.value,
      })
  }

  handleCommit() {
      $.post("/api/1/edit", {
          fileID: this.props.file.id,
          fileContent: this.state.fileContent,
      }).done(() => {
          window.location = "/edit/" + this.props.file.id
      })
  }

  handlePull() {
      $.post("/api/1/edit", {
          fileID: this.props.file.id,
          fileContent: this.state.fileContent,
      }).done(() => {
          window.location = "/edit/" + this.props.file.id
      })
  }

  render() {
      console.log(this.props, this.state);
      return <PagedrawGeneratedPage
          fileID={this.props.file.id}
          fileName={this.props.file.name}
          fileContent={this.state.fileContent}
          handleContentChange={this.handleContentChange.bind(this)}
          handleCommit={this.handleCommit.bind(this)}
          handlePull={handlePull}
          handlePush={handlePush}
      />;
  }
}

export default EditFile;
