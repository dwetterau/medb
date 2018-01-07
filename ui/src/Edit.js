import React, { Component } from 'react';
import PagedrawGeneratedPage from './pagedraw/editfile'
import $ from 'jquery';

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

  handleSave() {
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
          handleSave={this.handleSave.bind(this)}
      />;
  }
}

export default EditFile;
