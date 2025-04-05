import React from "react";

type FileSelectorProps = {
	onFileSelect: (file: File) => void;
};

const FileSelector: React.FC<FileSelectorProps> = ({ onFileSelect }) => {
	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const files = e.target.files;
		if (files && files.length > 0) {
			onFileSelect(files[0]);
		}
	};

	return <input type="file" accept="audio/*" onChange={handleChange} />;
};

export default FileSelector;
