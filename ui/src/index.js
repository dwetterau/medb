import React from 'react';
import ReactDOM from 'react-dom';
import $ from 'jquery';
import './index.css';
import App from './App';

// Read in the files.

$.getJSON('/api/1/list', (data) => {
    ReactDOM.render(<App rootFolderList={data}/>, document.getElementById('root'));
}).fail(function() {
    window.location.replace("/login.html")
});
