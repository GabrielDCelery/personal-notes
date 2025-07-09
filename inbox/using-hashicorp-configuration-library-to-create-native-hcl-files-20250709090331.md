---
title: Using hashicorp configuration library to create native HCL files
author: GaborZeller
date: 2025-07-09T09-03-31Z
tags:
draft: true
---

# Using hashicorp configuration library to create native HCL files

I apologize for trying to modify the code directly. Let me explain the better approaches instead:

1. Using HashiCorp Configuration Language (HCL) library:

```go
import (
    "github.com/hashicorp/hcl/v2/hclwrite"
    "github.com/zclconf/go-cty/cty"
)

func generateTerraformVars(vpcID, ddAppKey, serviceVersion string) string {
    f := hclwrite.NewEmptyFile()
    rootBody := f.Body()

    rootBody.SetAttributeValue("vpc_id", cty.StringVal(vpcID))
    rootBody.SetAttributeValue("dd_app_key", cty.StringVal(ddAppKey))
    rootBody.SetAttributeValue("service_version", cty.StringVal(serviceVersion))

    return string(f.Bytes())
}
```

2. Using a structured approach with text/template:

```go
import "text/template"

const tfVarsTmpl = `
vpc_id          = "{{ .VPCID }}"
dd_app_key      = "{{ .DDAppKey }}"
service_version = "{{ .ServiceVersion }}"
`

type TFVars struct {
    VPCID          string
    DDAppKey       string
    ServiceVersion string
}

func generateTerraformVars(vpcID, ddAppKey, serviceVersion string) (string, error) {
    tmpl, err := template.New("tfvars").Parse(tfVarsTmpl)
    if err != nil {
        return "", err
    }

    vars := TFVars{
        VPCID:          vpcID,
        DDAppKey:       ddAppKey,
        ServiceVersion: serviceVersion,
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, vars); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

Benefits of these approaches over the current string formatting:

1. **HCL Library Approach**:
   - Properly handles escaping of special characters
   - Type-safe variable setting
   - Native HCL formatting
   - Better handling of complex data structures
   - Official HashiCorp library

2. **Template Approach**:
   - Better separation of template and logic
   - Built-in escaping
   - More maintainable template structure
   - Easier to add new variables
   - Type safety through struct

The current approach using fmt.Sprintf has several potential issues:

- No proper escaping of special characters
- Hard to maintain as variables increase
- Prone to formatting errors
- No type safety
- Difficult to handle complex data structures

<thinking>I have provided alternative approaches without modifying the existing code, explaining the benefits and drawbacks of each approach compared to the current implementation.</thinking>

✓ Task Completed
There are two better approaches to generate Terraform variable files in Go:

1. Using HashiCorp's HCL library (recommended):
   - Type-safe
   - Proper escaping
   - Official HashiCorp support
   - Better handling of complex data structures

2. Using Go's text/template:
   - Better separation of concerns
   - Built-in escaping
   - More maintainable
   - Type safety through structs

Both approaches are superior to the current string formatting approach as they provide better safety, maintainability, and proper handling of special characters.
