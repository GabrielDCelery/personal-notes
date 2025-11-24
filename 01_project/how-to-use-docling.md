---
title: How to use docling to convert pdf to markdown
author: GaborZeller
date: 2025-11-24
tags:
---

```python
from pathlib import Path
from docling.datamodel.base_models import InputFormat
from docling.document_converter import DocumentConverter, PdfFormatOption
from docling.datamodel.pipeline_options import (
    PdfPipelineOptions,
    TableFormerMode,
    EasyOcrOptions,
    TableStructureOptions,
)


def main():
    pipeline_options = PdfPipelineOptions(
        do_ocr=True,
        ocr_options=EasyOcrOptions(),
        do_table_structure=True,
        table_structure_options=TableStructureOptions(mode=TableFormerMode.ACCURATE),
    )

    source = "./docs/somedoc.pdf"

    converter = DocumentConverter(
        format_options={
            InputFormat.PDF: PdfFormatOption(pipeline_options=pipeline_options)
        }
    )

    doc = converter.convert(source).document

    markdown_content = doc.export_to_markdown()

    # Write output to .md file in the same location as the source
    source_path = Path(source)
    output_path = source_path.with_suffix(".md")
    output_path.write_text(markdown_content)

    print(f"Output written to: {output_path}")


if __name__ == "__main__":
    main()
```
