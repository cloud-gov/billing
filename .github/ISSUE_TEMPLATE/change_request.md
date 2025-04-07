---
name: Change Request
about: Submit a change request for review and deployment.
title: "[Change Request] <Short Title>"
labels: change-request
assignees: ""
---

## Change Request Summary
Provide a concise summary of the change you are proposing. Include a brief title and description.

**Example:** "Upgrade authentication library to v2.5.1 to address CVE-2023-XXXX."

## Change Details
- **What is being changed?**  
  Describe the components or services affected.
  
- **Why is this change needed?**  
  Explain the motivation (bug fix, security patch, new feature, etc.).

- **Change Type:**  
  - [ ] Planned Change  
  - [ ] Emergency Change  

- **Priority:**  
  Indicate the appropriate priority level (P1, P2, P3, P4).

## Impact Analysis
- **Affected Systems/Components:**  
  List all systems or components impacted by this change.

- **Security Impact Analysis (SIA):**  
  - Describe potential security risks, vulnerabilities, or changes to controls.  
  - Has a SIA been performed? **Yes / No**  
  - Attach or link any supporting documentation.

- **Privacy Impact Assessment (PIA):**  
  - Describe any potential privacy risks or impacts on data handling.  
  - Has a PIA been performed? **Yes / No**  
  - Attach or link any supporting documentation.

## Rollback Plan
Outline your rollback plan if this change causes issues. Detail the steps to revert the change and restore service.

## Approvals
This change requires review and sign-off by the following stakeholders:
- [ ] **Security Team**  
- [ ] **Platform Operations**  
- [ ] **Pages Operations**  
- [ ] **System Owner**

*(If auto-assignment did not occur, please manually tag the appropriate reviewers.)*

## Additional Information
Provide any further details, related issue references, or links to documentation that may assist in the review process.

---

*Please ensure all sections are completed before submitting this change request.*
