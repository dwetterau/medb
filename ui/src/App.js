import React, { Component } from 'react';
import PagedrawGeneratedPage from './pagedraw/app'
import $ from 'jquery';

class App extends Component {
  constructor() {
      super();
      this.state = {
          filename: "",
          content: "",
      }
  }

  handleFilenameChange(e) {
       this.setState({
          filename: e.target.value,
          content: this.state.content,
      })
  }

  handleContentChange(e) {
      this.setState({
          filename: this.state.filename,
          content: e.target.value,
      })
  }

  handleSaveNote() {
      $.post("/api/1/save", {
          filename: this.state.filename,
          content: this.state.content,
      }).done(() => {
          // TODO: Something nicer on the refresh side?
          window.location = "/"
      })
  }

  render() {
    return (
        <PagedrawGeneratedPage
          rootFolderList={this.props.rootFolderList}
          filename={this.state.filename}
          content={this.state.content}
          handleFilenameChange={this.handleFilenameChange.bind(this)}
          handleContentChange={this.handleContentChange.bind(this)}
          handleSaveNote={this.handleSaveNote.bind(this)}
        />
    );
  }
}

export default App;
